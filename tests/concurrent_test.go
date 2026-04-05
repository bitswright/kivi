package tests

import (
	"fmt"
	"sync"
	"testing"

	"github.com/bitswright/kivi/internal/store"
)

func TestConcurrentWrites(t *testing.T) {
	s := store.NewMemStore()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", j)
			value := fmt.Sprintf("value-%d", j)
			if err := s.Set(key, value); err != nil {
				t.Errorf("Set failed: %v", err)
			}
		}(i)
	}

	wg.Wait()

	keys := s.Keys()
	if len(keys) != 100 {
		t.Errorf("expected 100 keys, got %d", len(keys))
	}
}

func TestConcurrentReadsAndWrites(t *testing.T) {
	s := store.NewMemStore()
	var wg sync.WaitGroup

	// seed one key first
	s.Set("shared", "initial")

	// 50 goroutines: reading
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Get("shared")
		}()
	}

	// 50 goroutines: writing
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			s.Set("shared", fmt.Sprintf("value-%d", j))
		}(i)
	}

	wg.Wait()
}
