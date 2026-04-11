package tests

import (
	"os"
	"testing"

	"github.com/bitswright/kivi/internal/store"
)

func TestWalstoreCrashRecovery(t *testing.T) {
	ws, err := store.NewWALStore("temp.wal")
	if err != nil {
		t.Fatalf("failed to create walstore: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Remove("temp.wal"); err != nil {
			t.Fatalf("failed to remove temp.wal: %v", err)
		}
	})

	// store a few keys
	testWalstoreSet("username", "alice", ws, t)
	testWalstoreSet("session_id", "abc123", ws, t)
	testWalstoreSet("theme", "light", ws, t)

	// do not close the ws, instead let it go out of scope
	// create new WALStore

	newWS, err := store.NewWALStore("temp.wal")
	if err != nil {
		t.Fatalf("failed to create new walstore: %v", err)
	}
	defer newWS.Close()

	// verify key are present
	testWalstoreGet("username", "alice", newWS, t)
	testWalstoreGet("session_id", "abc123", newWS, t)
	testWalstoreGet("theme", "light", newWS, t)
}

func testWalstoreSet(key, value string, ws *store.WALStore, t *testing.T) {
	if err := ws.Set(key, value); err != nil {
		t.Errorf("failed to set key %q: %v", key, err)
	}
}

func testWalstoreGet(key, expValue string, ws *store.WALStore, t *testing.T) {
	value, err := ws.Get(key)
	if err != nil {
		t.Errorf("failed to get key %q: %v", key, err)
		return
	}
	if value != expValue {
		t.Errorf("expected value %q for key %q, got %q", expValue, key, value)
	}
}
