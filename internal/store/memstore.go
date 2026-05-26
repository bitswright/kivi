package store

import (
	"sync"
	"time"
)

type entry[V any] struct {
	value     V
	expiresAt time.Time
}

// returns true if the entry has an expiration time and it has passed.
func (e entry[V]) isExpired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}

type MemStore[V any] struct {
	mu   sync.RWMutex
	data map[string]entry[V]
}

// creates and returns an empty MemStore
func New[V any]() *MemStore[V] {
	return &MemStore[V]{
		data: make(map[string]entry[V]),
	}
}

// Set stores the given value under given key with no expiry
func (s *MemStore[V]) Set(key string, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = entry[V]{value: value}
}

// Set stores the given value under given key that expires after ttl duration
func (s *MemStore[V]) SetWithTTL(key string, value V, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = entry[V]{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Get returns the value for key and whether it was found.
// Returns false if the key does not exist or has expired.
func (s *MemStore[V]) Get(key string) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.data[key]
	if !ok || e.isExpired() {
		var zero V
		// zero will be initialized with zero-value based on V type
		return zero, false
	}
	return e.value, true
}

func (s *MemStore[V]) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
}

// Len returns the number of keys currently in the store,
// including expired keys not yet reaped.
func (m *MemStore[V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.data)
}
