package store

import (
	"sync"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	m := New[string]()

	m.Set("name", "kivi")

	val, ok := m.Get("name")

	if !ok {
		t.Fatal("expected key to exist")
	}
	if val != "kivi" {
		t.Fatalf("expected 'kivi', got %q", val)
	}
}

func TestGetMissingKey(t *testing.T) {
	m := New[string]()

	_, ok := m.Get("ghost")
	if ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestDelete(t *testing.T) {
	m := New[int]()

	m.Set("version", 4)
	m.Delete("version")

	_, ok := m.Get("version")
	if ok {
		t.Fatal("expected deleted key to return false")
	}
}

func TestTTLExpiry(t *testing.T) {
	m := New[string]()

	m.SetWithTTL("temp", "value", 2*time.Second)

	// should exist immediately
	_, ok := m.Get("temp")
	if !ok {
		t.Fatal("expected key to exist before expiry")
	}

	// should be gone after TTL
	time.Sleep(3 * time.Second)
	_, ok = m.Get("temp")
	if ok {
		t.Fatal("expected key to not exist after expiry")
	}
}

func TestConcurrentReadsDoNotBlock(t *testing.T) {
	m := New[int]()
	for i := range 100 {
		m.Set(string(rune('a'+i%26)), i)
	}

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Get("a")
			m.Get("b")
		}()
	}
	wg.Wait() //if this deadlocks, concurrent reads are broken
}
