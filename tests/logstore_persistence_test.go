package tests

import (
	"os"
	"testing"

	"github.com/bitswright/kivi/internal/store"
)

func TestLogStorePersistence(t *testing.T) {
	// create a log store pointing to a temp file
	ls, err := store.NewLogStore("temp.log")
	if err != nil {
		t.Fatalf("Logstore initialization failed: %v", err)
	}
	t.Cleanup(func() {
		os.Remove("temp.log")
	})

	// set a few keys
	testLogStoreSet("username", "john", ls, t)
	testLogStoreSet("session_id", "123", ls, t)
	testLogStoreSet("theme", "dark", ls, t)

	// close the store
	if err := ls.Close(); err != nil {
		t.Fatalf("Logstore closing failed: %v", err)
	}

	// create a new log store pointing to same file
	ls, err = store.NewLogStore("temp.log")
	if err != nil {
		t.Fatalf("Logstore re-initialization failed: %v", err)
	}

	// check keys persistence
	testLogStoreGet("username", "john", ls, t)
	testLogStoreGet("session_id", "123", ls, t)
	testLogStoreGet("theme", "dark", ls, t)
}

func testLogStoreSet(key, value string, ls *store.LogStore, t *testing.T) {
	if err := ls.Set(key, value); err != nil {
		t.Errorf("Set failed: %v", err)
	}
}

func testLogStoreGet(key, expValue string, ls *store.LogStore, t *testing.T) {
	actualValue, err := ls.Get(key)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}

	if actualValue != expValue {
		t.Errorf("expected: %s, got %s", expValue, actualValue)
	}
}
