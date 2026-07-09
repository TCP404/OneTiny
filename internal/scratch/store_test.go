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

func TestStoreListReturnsCopy(t *testing.T) {
	store, err := NewStore(Limits{MaxItems: 3, MaxItemBytes: 1024})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	_, err = store.Add(KindText, TextMIME, []byte("hello"))
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	items := store.List()
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	items[0].Data[0] = 'H'

	got, ok := store.Get(ContentID(KindText, TextMIME, []byte("hello")))
	if !ok {
		t.Fatal("Get returned not found for existing item")
	}
	if string(got.Data) != "hello" {
		t.Fatalf("stored data mutated via List result: %q", string(got.Data))
	}
}

func TestStoreGetReturnsCopy(t *testing.T) {
	store, err := NewStore(Limits{MaxItems: 3, MaxItemBytes: 1024})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	added, err := store.Add(KindText, TextMIME, []byte("hello"))
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	got, ok := store.Get(added.ID)
	if !ok {
		t.Fatal("Get returned not found for existing item")
	}
	got.Data[0] = 'H'

	again, ok := store.Get(added.ID)
	if !ok {
		t.Fatal("second Get returned not found for existing item")
	}
	if string(again.Data) != "hello" {
		t.Fatalf("stored data mutated via Get result: %q", string(again.Data))
	}
}

func TestStoreListSummariesReturnsLimitedTextPreview(t *testing.T) {
	store, err := NewStore(Limits{MaxItems: 3, MaxItemBytes: 1024})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	data := []byte("abcdef")
	added, err := store.Add(KindText, TextMIME, data)
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	summaries := store.ListSummaries(4)
	if len(summaries) != 1 {
		t.Fatalf("len(summaries) = %d, want 1", len(summaries))
	}
	summary := summaries[0]
	if summary.ID != added.ID {
		t.Fatalf("summary ID = %q, want %q", summary.ID, added.ID)
	}
	if summary.Size != len(data) {
		t.Fatalf("summary size = %d, want %d", summary.Size, len(data))
	}
	if string(summary.Preview) != "abcd" {
		t.Fatalf("summary preview = %q, want abcd", string(summary.Preview))
	}
	if !summary.Truncated {
		t.Fatal("summary should report truncated preview")
	}

	summary.Preview[0] = 'A'
	got, ok := store.Get(added.ID)
	if !ok {
		t.Fatal("Get returned not found for existing item")
	}
	if string(got.Data) != string(data) {
		t.Fatalf("stored data mutated via summary preview: %q", string(got.Data))
	}
}

func TestStoreConcurrentAddAndList(t *testing.T) {
	store, err := NewStore(Limits{MaxItems: 200, MaxItemBytes: 1024})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	expected := make(map[string][]byte, 50)
	var expectedMu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			payload := bytes.Repeat([]byte{byte(i)}, 8)
			added, err := store.Add(KindText, TextMIME, payload)
			if err != nil {
				t.Errorf("Add(%d) returned error: %v", i, err)
				return
			}
			expectedMu.Lock()
			expected[added.ID] = append([]byte(nil), payload...)
			expectedMu.Unlock()
			_ = store.List()
		}(i)
	}
	wg.Wait()

	items := store.List()
	if len(items) != 50 {
		t.Fatalf("len(items) = %d, want 50", len(items))
	}
	for _, item := range items {
		expectedMu.Lock()
		wantData, ok := expected[item.ID]
		expectedMu.Unlock()
		if !ok {
			t.Fatalf("unexpected item ID %q", item.ID)
		}
		if _, ok := store.Get(item.ID); !ok {
			t.Fatalf("Get(%q) reported missing item", item.ID)
		}
		got, ok := store.Get(item.ID)
		if !ok {
			t.Fatalf("Get(%q) returned missing item on second read", item.ID)
		}
		if !bytes.Equal(got.Data, wantData) {
			t.Fatalf("Get(%q) data = %v, want %v", item.ID, got.Data, wantData)
		}
		if item.Kind != KindText || item.MimeType != TextMIME {
			t.Fatalf("item = %+v, want text/plain", item)
		}
	}

	for id := range expected {
		expectedMu.Lock()
		wantData := expected[id]
		expectedMu.Unlock()
		item, ok := store.Get(id)
		if !ok {
			t.Fatalf("expected item %q not found", id)
		}
		if !bytes.Equal(item.Data, wantData) {
			t.Fatalf("item %q data corrupted: %v", id, item.Data)
		}
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
