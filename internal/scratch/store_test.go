package scratch

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

func TestStoreAddsTextAndUsesContentHashID(t *testing.T) {
	store, err := NewStore(Limits{MaxItems: 5, MaxItemBytes: 1024})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	item, err := store.Add(KindText, TextMIME, []byte("hello"))
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if item.ID != ContentID(KindText, TextMIME, []byte("hello")) {
		t.Fatalf("item ID = %q, want content hash", item.ID)
	}
	if item.Kind != KindText || item.MimeType != TextMIME {
		t.Fatalf("item type = %s %s, want text mime", item.Kind, item.MimeType)
	}
	if string(item.Data) != "hello" {
		t.Fatalf("item data = %q, want hello", string(item.Data))
	}
}

func TestStoreAddsSupportedImage(t *testing.T) {
	store, err := NewStore(Limits{MaxItems: 5, MaxItemBytes: 1024})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	item, err := store.Add(KindImage, "image/png", []byte{0x89, 0x50, 0x4e, 0x47})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if item.Kind != KindImage || item.MimeType != "image/png" {
		t.Fatalf("item = %+v, want png image", item)
	}
}

func TestDuplicateContentMovesExistingItemToTop(t *testing.T) {
	store, err := NewStore(Limits{MaxItems: 10, MaxItemBytes: 1024})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	store.now = sequenceClock(
		time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 8, 10, 1, 0, 0, time.UTC),
		time.Date(2026, 7, 8, 10, 2, 0, 0, time.UTC),
	)

	first, _ := store.Add(KindText, TextMIME, []byte("1"))
	second, _ := store.Add(KindText, TextMIME, []byte("2"))
	touched, err := store.Add(KindText, TextMIME, []byte("1"))
	if err != nil {
		t.Fatalf("duplicate Add returned error: %v", err)
	}

	items := store.List()
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].ID != first.ID || touched.ID != first.ID || items[1].ID != second.ID {
		t.Fatalf("order = %v, want duplicate item first", ids(items))
	}
	if !items[0].UpdatedAt.After(items[0].CreatedAt) {
		t.Fatalf("UpdatedAt = %v should be after CreatedAt = %v", items[0].UpdatedAt, items[0].CreatedAt)
	}
}

func TestStoreTrimsOldestItemAtCapacity(t *testing.T) {
	store, err := NewStore(Limits{MaxItems: 3, MaxItemBytes: 1024})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	_, _ = store.Add(KindText, TextMIME, []byte("1"))
	_, _ = store.Add(KindText, TextMIME, []byte("2"))
	_, _ = store.Add(KindText, TextMIME, []byte("3"))
	_, _ = store.Add(KindText, TextMIME, []byte("4"))

	got := ids(store.List())
	want := []string{
		ContentID(KindText, TextMIME, []byte("4")),
		ContentID(KindText, TextMIME, []byte("3")),
		ContentID(KindText, TextMIME, []byte("2")),
	}
	if !equalStrings(got, want) {
		t.Fatalf("ids = %v, want %v", got, want)
	}
	if _, ok := store.Get(ContentID(KindText, TextMIME, []byte("1"))); ok {
		t.Fatal("oldest item still exists after capacity trim")
	}
}

func TestStoreRejectsOversizedItemWithoutChangingList(t *testing.T) {
	store, err := NewStore(Limits{MaxItems: 3, MaxItemBytes: 3})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	_, _ = store.Add(KindText, TextMIME, []byte("ok"))

	_, err = store.Add(KindText, TextMIME, []byte("toolong"))
	if err != ErrItemTooLarge {
		t.Fatalf("Add error = %v, want %v", err, ErrItemTooLarge)
	}
	if len(store.List()) != 1 {
		t.Fatalf("list changed after oversized add: %+v", store.List())
	}
}

func TestStoreRejectsUnsupportedMime(t *testing.T) {
	store, err := NewStore(Limits{MaxItems: 3, MaxItemBytes: 1024})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	_, err = store.Add(KindImage, "image/bmp", []byte("bmp"))
	if err != ErrUnsupportedType {
		t.Fatalf("Add error = %v, want %v", err, ErrUnsupportedType)
	}
}

func TestStoreConcurrentAddAndList(t *testing.T) {
	store, err := NewStore(Limits{MaxItems: 200, MaxItemBytes: 1024})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			payload := bytes.Repeat([]byte{byte(i)}, 8)
			_, _ = store.Add(KindText, TextMIME, payload)
			_ = store.List()
		}(i)
	}
	wg.Wait()

	if len(store.List()) == 0 {
		t.Fatal("concurrent add produced empty list")
	}
}

func sequenceClock(values ...time.Time) func() time.Time {
	index := 0
	return func() time.Time {
		if index >= len(values) {
			return values[len(values)-1]
		}
		value := values[index]
		index++
		return value
	}
}

func ids(items []Item) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.ID
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
