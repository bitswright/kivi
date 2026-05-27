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
	for i := range 26 {
		m.Set(string(rune('a'+i)), i)
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

func TestConcurrentWritesDoNotCorrupt(t *testing.T) {
	m := New[int]()

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Set(string(rune('a'+i%26)), i)
		}()
	}
	wg.Wait()
	// if we get here without panic, map wasn't corrupted
}

func TestScanPrefix(t *testing.T) {
	m := New[int]()

	m.Set("bar", 1)
	m.Set("foo", 2)
	m.Set("foo1", 3)
	m.Set("bar2", 4)
	m.Set("bar3", 5)
	m.Set("foo2", 6)

	var got []string
	for k := range m.Scan("foo") {
		got = append(got, k)
	}

	expected := []string{"foo", "foo1", "foo2"}
	if len(got) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}

	for i, k := range got {
		if k != expected[i] {
			t.Fatalf("at index %d: expected %q, got %q", i, expected[i], k)
		}
	}
}

func TestScanSorted(t *testing.T) {
	m := New[int]()

	m.Set("key:b", 2)
	m.Set("key:c", 3)
	m.Set("key:a", 1)

	var got []string
	for k := range m.Scan("key") {
		got = append(got, k)
	}

	expected := []string{"key:a", "key:b", "key:c"}
	if len(got) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}

	for i, k := range got {
		if k != expected[i] {
			t.Fatalf("at index %d: expected %q, got %q", i, expected[i], k)
		}
	}
}

func TestScanEmptyPrefix(t *testing.T) {
	m := New[int]()

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	var got []string
	for k := range m.Scan("") {
		got = append(got, k)
	}

	expected := []string{"a", "b", "c"}
	if len(got) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}

	for i, k := range got {
		if k != expected[i] {
			t.Fatalf("at index %d: expected %q, got %q", i, expected[i], k)
		}
	}

}

func TestScanEarlyExit(t *testing.T) {
	m := New[int]()

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	var got []string
	for k := range m.Scan("") {
		got = append(got, k)
		if len(got) == 2 {
			break
		}
	}

	expected := []string{"a", "b"}
	if len(got) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}

	for i, k := range got {
		if k != expected[i] {
			t.Fatalf("at index %d: expected %q, got %q", i, expected[i], k)
		}
	}
}

func TestScanExcludesExpired(t *testing.T) {
	m := New[int]()

	m.Set("a", 1)
	m.SetWithTTL("b", 2, 1*time.Millisecond)
	m.Set("c", 3)

	time.Sleep(5 * time.Millisecond)

	var got []string
	for k := range m.Scan("") {
		got = append(got, k)
	}

	expected := []string{"a", "c"}
	if len(got) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}

	for i, k := range got {
		if k != expected[i] {
			t.Fatalf("at index %d: expected %q, got %q", i, expected[i], k)
		}
	}
}
