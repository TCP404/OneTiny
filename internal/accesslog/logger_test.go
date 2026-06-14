package accesslog

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLoggerWriteReadFilterAndClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "access.log")
	logger := New(path)

	base := time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC)
	events := []Event{
		{Time: base, Event: EventAccess, Method: "GET", Path: "/ok", Status: 200, Result: ResultSuccess},
		{Time: base.Add(time.Minute), Event: EventDownload, Method: "GET", Path: "/file/a.txt", Status: 200, Result: ResultSuccess},
		{Time: base.Add(2 * time.Minute), Event: EventUpload, Method: "POST", Path: "/file/upload", Status: 500, Result: ResultFailure},
	}
	for _, event := range events {
		if err := logger.Write(event); err != nil {
			t.Fatalf("Write(%+v) returned error: %v", event, err)
		}
	}

	got, err := logger.Read(Filter{})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(got) != len(events) {
		t.Fatalf("Read returned %d events, want %d", len(got), len(events))
	}

	downloads, err := logger.Read(Filter{Event: EventDownload})
	if err != nil {
		t.Fatalf("Read download filter returned error: %v", err)
	}
	if len(downloads) != 1 || downloads[0].Path != "/file/a.txt" {
		t.Fatalf("download filter = %+v, want one download event", downloads)
	}

	window, err := logger.Read(Filter{Since: base.Add(30 * time.Second), Until: base.Add(90 * time.Second)})
	if err != nil {
		t.Fatalf("Read time filter returned error: %v", err)
	}
	if len(window) != 1 || window[0].Event != EventDownload {
		t.Fatalf("time filter = %+v, want download event only", window)
	}

	if err := logger.Clear(); err != nil {
		t.Fatalf("Clear returned error: %v", err)
	}
	afterClear, err := logger.Read(Filter{})
	if err != nil {
		t.Fatalf("Read after Clear returned error: %v", err)
	}
	if len(afterClear) != 0 {
		t.Fatalf("Read after Clear returned %d events, want 0", len(afterClear))
	}
}

func TestLoggerReadMissingFileReturnsEmptySlice(t *testing.T) {
	logger := New(filepath.Join(t.TempDir(), "missing.log"))

	got, err := logger.Read(Filter{})
	if err != nil {
		t.Fatalf("Read missing file returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Read missing file returned %d events, want 0", len(got))
	}
}

func TestLoggerReadSkipsMalformedLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "access.log")
	if err := os.WriteFile(path, []byte("{bad json\n{\"event\":\"access\",\"result\":\"success\"}\n"), 0o600); err != nil {
		t.Fatalf("write malformed log: %v", err)
	}
	logger := New(path)

	got, err := logger.Read(Filter{})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(got) != 1 || got[0].Event != EventAccess {
		t.Fatalf("Read malformed log = %+v, want one valid access event", got)
	}
}

func TestLoggerReadSkipsOversizedMalformedLineAndContinues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "access.log")
	longBadLine := strings.Repeat("x", 2*1024*1024)
	validLine := `{"event":"access","result":"success","path":"/after-long-line"}`
	if err := os.WriteFile(path, []byte(longBadLine+"\n"+validLine+"\n"), 0o600); err != nil {
		t.Fatalf("write oversized malformed log: %v", err)
	}
	logger := New(path)

	got, err := logger.Read(Filter{})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(got) != 1 || got[0].Path != "/after-long-line" {
		t.Fatalf("Read oversized malformed log = %+v, want one valid event after long line", got)
	}
}

func TestLoggerWriteFillsEmptyTime(t *testing.T) {
	logger := New(filepath.Join(t.TempDir(), "access.log"))

	if err := logger.Write(Event{Event: EventLogin, Result: ResultSuccess}); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	got, err := logger.Read(Filter{})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Read returned %d events, want 1", len(got))
	}
	if got[0].Time.IsZero() {
		t.Fatalf("Write left event time empty: %+v", got[0])
	}
}

func TestLoggerConcurrentWritesDoNotInterleaveJSONLines(t *testing.T) {
	logger := New(filepath.Join(t.TempDir(), "access.log"))

	const count = 64
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			if err := logger.Write(Event{Event: EventAccess, Result: ResultSuccess, Status: 200}); err != nil {
				t.Errorf("Write returned error: %v", err)
			}
		}()
	}
	wg.Wait()

	got, err := logger.Read(Filter{Event: EventAccess})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(got) != count {
		t.Fatalf("Read returned %d events, want %d", len(got), count)
	}
}
