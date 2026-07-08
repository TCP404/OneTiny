package scratch

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

type Kind string

const (
	KindText  Kind = "text"
	KindImage Kind = "image"

	TextMIME = "text/plain; charset=utf-8"
)

var (
	ErrInvalidLimits   = errors.New("invalid limits")
	ErrEmptyContent    = errors.New("empty content")
	ErrItemTooLarge    = errors.New("item too large")
	ErrUnsupportedType = errors.New("unsupported type")
	ErrItemNotFound    = errors.New("item not found")
)

type Limits struct {
	MaxItems     int
	MaxItemBytes int
}

type Item struct {
	ID        string
	Kind      Kind
	MimeType  string
	Data      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Store struct {
	mu     sync.RWMutex
	now    func() time.Time
	limits Limits
	items  []*Item
	index  map[string]int
}

func NewStore(l Limits) (*Store, error) {
	if err := validateLimits(l); err != nil {
		return nil, err
	}
	return &Store{
		now:    time.Now,
		limits: l,
		items:  make([]*Item, 0, l.MaxItems),
		index:  make(map[string]int),
	}, nil
}

func (s *Store) UpdateLimits(l Limits) error {
	if err := validateLimits(l); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.limits = l
	s.trimToCapacityLocked()
	return nil
}

func (s *Store) Limits() Limits {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.limits
}

func (s *Store) Add(kind Kind, mimeType string, data []byte) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(data) == 0 {
		return Item{}, ErrEmptyContent
	}
	if len(data) > s.limits.MaxItemBytes {
		return Item{}, ErrItemTooLarge
	}
	if !isSupportedType(kind, mimeType) {
		return Item{}, ErrUnsupportedType
	}

	id := ContentID(kind, mimeType, data)
	now := s.currentTimeLocked()
	copyData := append([]byte(nil), data...)

	if existingIdx, ok := s.index[id]; ok {
		existing := s.items[existingIdx]
		existing.Kind = kind
		existing.MimeType = mimeType
		existing.Data = copyData
		existing.UpdatedAt = now
		s.moveToTopLocked(existingIdx)
		s.trimToCapacityLocked()
		return cloneItem(existing), nil
	}

	item := &Item{
		ID:        id,
		Kind:      kind,
		MimeType:  mimeType,
		Data:      copyData,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.items = append([]*Item{item}, s.items...)
	s.reindexLocked()
	s.trimToCapacityLocked()
	return cloneItem(item), nil
}

func (s *Store) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Item, len(s.items))
	for i, item := range s.items {
		out[i] = cloneItem(item)
	}
	return out
}

func (s *Store) Get(id string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idx, ok := s.index[id]
	if !ok {
		return Item{}, false
	}
	return cloneItem(s.items[idx]), true
}

func ContentID(kind Kind, mimeType string, data []byte) string {
	sum := sha256.Sum256(append(append(append([]byte(string(kind)), 0), []byte(mimeType)...), append([]byte{0}, data...)...))
	return hex.EncodeToString(sum[:])
}

func validateLimits(l Limits) error {
	if l.MaxItems <= 0 || l.MaxItemBytes <= 0 {
		return ErrInvalidLimits
	}
	return nil
}

func isSupportedType(kind Kind, mimeType string) bool {
	switch kind {
	case KindText:
		return mimeType == TextMIME
	case KindImage:
		switch mimeType {
		case "image/png", "image/jpeg", "image/gif", "image/webp":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func (s *Store) currentTimeLocked() time.Time {
	now := s.now
	if now == nil {
		return time.Now()
	}
	return now()
}

func (s *Store) moveToTopLocked(idx int) {
	if idx <= 0 || idx >= len(s.items) {
		if idx == 0 {
			s.reindexLocked()
		}
		return
	}
	item := s.items[idx]
	copy(s.items[1:idx+1], s.items[0:idx])
	s.items[0] = item
	s.reindexLocked()
}

func (s *Store) trimToCapacityLocked() {
	if s.limits.MaxItems <= 0 {
		return
	}
	if len(s.items) <= s.limits.MaxItems {
		s.reindexLocked()
		return
	}
	s.items = append([]*Item(nil), s.items[:s.limits.MaxItems]...)
	s.reindexLocked()
}

func (s *Store) reindexLocked() {
	s.index = make(map[string]int, len(s.items))
	for i, item := range s.items {
		s.index[item.ID] = i
	}
}

func cloneItem(item *Item) Item {
	if item == nil {
		return Item{}
	}
	out := *item
	out.Data = append([]byte(nil), item.Data...)
	return out
}
