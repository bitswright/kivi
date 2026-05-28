package store

import (
	"iter"
	"sort"
	"strings"
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
	mu       sync.RWMutex
	data     map[string]entry[V]
	done     chan struct{}
	stopOnce sync.Once
}

// creates and returns an empty MemStore
func New[V any](reaperInterval time.Duration) *MemStore[V] {
	s := &MemStore[V]{
		data: make(map[string]entry[V]),
		done: make(chan struct{}),
	}
	go s.reap(reaperInterval)
	return s
}

// reap runs on a ticker and deletes expired keys from the map.
func (s *MemStore[V]) reap(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.deleteExpired()
		case <-s.done:
			return
		}
	}
}

// deleteExpired removes all expired entries under a write lock.
func (s *MemStore[V]) deleteExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, e := range s.data {
		if e.isExpired() {
			delete(s.data, k)
		}
	}
}

// Stop shuts down the background reaper goroutine.
// Safe to call multiple times.
func (s *MemStore[V]) Stop() {
	s.stopOnce.Do(func() {
		close(s.done)
	})
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
func (s *MemStore[V]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.data)
}

func (s *MemStore[V]) Scan(prefix string) iter.Seq2[string, V] {
	return func(yield func(string, V) bool) {
		s.mu.RLock()
		keys := make([]string, 0)
		for k, e := range s.data {
			if e.isExpired() {
				continue
			}
			if strings.HasPrefix(k, prefix) {
				keys = append(keys, k)
			}
		}
		s.mu.RUnlock()

		// no need to keep lock while sorting the keys
		sort.Strings(keys)

		for _, k := range keys {
			s.mu.RLock()
			e, ok := s.data[k]
			s.mu.RUnlock()

			if !ok || e.isExpired() {
				continue // key got deleted or expired
			}

			if !yield(k, e.value) {
				return // caller broke out of loop early
			}
		}
	}
}
