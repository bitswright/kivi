package memstore

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func BenchmarkGet(b *testing.B) {
	m := New[[]byte](time.Minute)
	defer m.Stop()

	m.Set("key", []byte("value"))

	b.ResetTimer()

	for range b.N {
		m.Get("key")
	}
}

func BenchmarkSet(b *testing.B) {
	m := New[[]byte](time.Minute)
	defer m.Stop()

	b.ResetTimer()

	for i := range b.N {
		m.Set(fmt.Sprintf("key-%d", i), []byte("value"))
	}
}

// BenchmarkGetParallel measures read performance under real concurrency.
func BenchmarkGetParallel(b *testing.B) {
	m := New[[]byte](time.Minute)
	defer m.Stop()

	m.Set("key", []byte("value"))

	b.ResetTimer()

	// b.RunParallel spins up GOMAXPROCS goroutines all hitting the store simultaneously.
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Get("key")
		}
	})
}

// BenchmarkConcurrentRun measures throughput when many goroutines read different keys simultaneously
func BenchmarkConcurrentRun(b *testing.B) {
	m := New[[]byte](time.Minute)
	defer m.Stop()

	// pre-populate 1000 keys
	for i := range 1000 {
		m.Set(fmt.Sprintf("key-%d", i), []byte("value"))
	}

	b.ResetTimer()

	var wg sync.WaitGroup
	for range b.N {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Get("key-500")
		}()
	}
	wg.Wait()
}

// BenchmarkSetParallel measures write throughput under real concurrency.
func BenchmarkSetParallel(b *testing.B) {
	m := New[[]byte](time.Minute)
	defer m.Stop()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Set(fmt.Sprintf("key-%d", i), []byte("value"))
			i++
		}
	})
}

// BenchmarkScan measures prefix scan over 1000 keys.
func BenchmarkScan(b *testing.B) {
	m := New[[]byte](time.Minute)
	defer m.Stop()

	for i := range 1000 {
		m.Set(fmt.Sprintf("foo:%d", i), []byte("value"))
	}

	b.ResetTimer()

	for range b.N {
		for range m.Scan("foo") {
			// just iterate, don't process
		}
	}
}
