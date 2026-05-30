package db

import (
	"path/filepath"
	"testing"

	"github.com/bitswright/kivi/internal/logstore"
)

func tempDB(t *testing.T) *DB {
	t.Helper()
	opts := Options{
		LogPath:  filepath.Join(t.TempDir(), "kivi.log"),
		SyncMode: logstore.SyncNone,
		// no reaper needed for these tests
		// zero ReaperInterval would panic on NewTicker
		// max duration — effectively never ticks
		ReaperInterval: 1<<63 - 1,
	}

	db, err := Open(opts)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	t.Cleanup(func() { db.Close() })

	return db
}

func TestSetAndGet(t *testing.T) {
	db := tempDB(t)

	if err := db.Set("name", []byte("kivi")); err != nil {
		t.Fatal(err)
	}

	val, ok := db.Get("name")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if string(val) != "kivi" {
		t.Fatalf("expected 'kivi', got %q", val)
	}
}

func TestDelete(t *testing.T) {
	db := tempDB(t)

	db.Set("x", []byte("hello"))
	db.Delete("x")

	_, ok := db.Get("x")
	if ok {
		t.Fatal("expected deleted key to be gone")
	}
}

func TestSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "kivi.log")

	opts := Options{
		LogPath:        logPath,
		SyncMode:       logstore.SyncFull,
		ReaperInterval: 1<<63 - 1,
	}

	// first session — write some data
	db1, err := Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	db1.Set("foo", []byte("bar"))
	db1.Set("baz", []byte("qux"))
	db1.Delete("foo")
	db1.Close()

	// second session — reopen and verify state was reconstructed
	db2, err := Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()

	// "foo" was deleted — should not exist
	_, ok := db2.Get("foo")
	if ok {
		t.Fatal("expected 'foo' to be deleted after restart")
	}

	// "baz" was set — should exist
	val, ok := db2.Get("baz")
	if !ok {
		t.Fatal("expected 'baz' to exist after restart")
	}
	if string(val) != "qux" {
		t.Fatalf("expected 'qux', got %q", val)
	}
}

func TestOverwriteSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "kivi.log")

	opts := Options{
		LogPath:        logPath,
		SyncMode:       logstore.SyncFull,
		ReaperInterval: 1<<63 - 1,
	}

	db1, _ := Open(opts)
	db1.Set("key", []byte("first"))
	db1.Set("key", []byte("second")) // overwrite
	db1.Close()

	db2, _ := Open(opts)
	defer db2.Close()

	val, ok := db2.Get("key")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if string(val) != "second" {
		t.Fatalf("expected 'second', got %q", val)
	}
}

func TestScanAfterRestart(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "kivi.log")

	opts := Options{
		LogPath:        logPath,
		SyncMode:       logstore.SyncNone,
		ReaperInterval: 1<<63 - 1,
	}

	db1, _ := Open(opts)
	db1.Set("user:1", []byte("alice"))
	db1.Set("user:2", []byte("bob"))
	db1.Set("user:3", []byte("carol"))
	db1.Close()

	db2, _ := Open(opts)
	defer db2.Close()

	var keys []string
	for k := range db2.Scan("user") {
		keys = append(keys, k)
	}

	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %v", keys)
	}

	// verify sorted order
	expected := []string{"user:1", "user:2", "user:3"}
	for i, k := range keys {
		if k != expected[i] {
			t.Fatalf("index %d: expected %q, got %q", i, expected[i], k)
		}
	}
}
