package accesslog

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	maxLogLineBytes = 1024 * 1024

	EventAccess   = "access"
	EventDownload = "download"
	EventUpload   = "upload"
	EventLogin    = "login"
	EventReject   = "reject"
	EventError    = "error"

	ResultSuccess = "success"
	ResultFailure = "failure"
	ResultReject  = "reject"
)

type Event struct {
	Time     time.Time `json:"time"`
	ClientIP string    `json:"client_ip"`
	Method   string    `json:"method"`
	Event    string    `json:"event"`
	Path     string    `json:"path"`
	Status   int       `json:"status"`
	Result   string    `json:"result"`
}

type Filter struct {
	Event string
	Since time.Time
	Until time.Time
}

type Logger struct {
	path string
	mu   sync.Mutex
}

var (
	defaultLogger   = New(DefaultPath())
	defaultLoggerMu sync.RWMutex
)

func DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "./access.log"
	}
	return filepath.Join(dir, "tiny", "access.log")
}

func New(path string) *Logger {
	return &Logger{path: path}
}

func SetLoggerForTest(logger *Logger) func() {
	defaultLoggerMu.Lock()
	previous := defaultLogger
	defaultLogger = logger
	defaultLoggerMu.Unlock()

	return func() {
		defaultLoggerMu.Lock()
		defaultLogger = previous
		defaultLoggerMu.Unlock()
	}
}

func Log(event Event) {
	logger := currentLogger()
	if logger == nil {
		return
	}
	_ = logger.Write(event)
}

func (l *Logger) Write(event Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if event.Time.IsZero() {
		event.Time = time.Now()
	}
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	return json.NewEncoder(file).Encode(event)
}

func (l *Logger) Read(filter Filter) ([]Event, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	file, err := os.Open(l.path)
	if errors.Is(err, os.ErrNotExist) {
		return []Event{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	events := make([]Event, 0)
	reader := bufio.NewReader(file)
	for {
		line, err := readLogLine(reader)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var event Event
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}
		if !matchesFilter(event, filter) {
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

func (l *Logger) Clear() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(l.path, nil, 0o600)
}

func currentLogger() *Logger {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	return defaultLogger
}

func readLogLine(reader *bufio.Reader) ([]byte, error) {
	var line []byte
	oversized := false

	for {
		fragment, err := reader.ReadSlice('\n')
		if len(fragment) > 0 {
			if !oversized && len(line)+len(fragment) <= maxLogLineBytes {
				line = append(line, fragment...)
			} else {
				oversized = true
			}
		}

		switch {
		case err == nil:
			if oversized {
				return nil, nil
			}
			return line, nil
		case errors.Is(err, bufio.ErrBufferFull):
			continue
		case errors.Is(err, io.EOF):
			if oversized || len(line) == 0 {
				return nil, io.EOF
			}
			return line, nil
		default:
			return nil, err
		}
	}
}

func matchesFilter(event Event, filter Filter) bool {
	if filter.Event != "" && event.Event != filter.Event {
		return false
	}
	if !filter.Since.IsZero() && event.Time.Before(filter.Since) {
		return false
	}
	if !filter.Until.IsZero() && event.Time.After(filter.Until) {
		return false
	}
	return true
}
