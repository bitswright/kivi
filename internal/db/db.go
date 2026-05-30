package db

import (
	"fmt"
	"iter"
	"time"

	"github.com/bitswright/kivi/internal/logstore"
	"github.com/bitswright/kivi/internal/memstore"
)

// Options configures the DB on open.
type Options struct {
	// Path to the log file on disk.
	LogPath string

	// SyncMode controls durability vs performance trade-off.
	SyncMode logstore.SyncMode

	// ReaperInterval controls how often expired keys are cleaned up.
	ReaperInterval time.Duration
}

// DefaultOptions returns sensible defaults for development.
func DefaultOptions(logPath string) Options {
	return Options{
		LogPath:        logPath,
		SyncMode:       logstore.SyncFull,
		ReaperInterval: 30 * time.Second,
	}
}

// DB is a persistent key-value store.
// It combines an in-memory store with an append-only log for durability.
type DB struct {
	mem *memstore.MemStore[[]byte]
	log *logstore.LogStore
}

// Open opens or creates a DB at the given path.
// If a log file exists, it is replayed to reconstruct in-memory state.
func Open(opts Options) (*DB, error) {
	// open the log file
	l, err := logstore.Open(opts.LogPath, opts.SyncMode)
	if err != nil {
		return nil, fmt.Errorf("open logstore: %w", err)
	}

	// create empty memstore
	mem := memstore.New[[]byte](opts.ReaperInterval)

	// replay log into memstore to reconstruct state
	ReplayOperation := func(rec logstore.Record) error {
		switch rec.Type {
		case logstore.RecordPut:
			mem.Set(string(rec.Key), rec.Value)
		case logstore.RecordDelete:
			mem.Delete(string(rec.Key))
		}
		return nil
	}
	if err := l.Replay(ReplayOperation); err != nil {
		mem.Stop()
		l.Close()
		return nil, fmt.Errorf("replay log: %w", err)
	}

	return &DB{mem: mem, log: l}, nil
}

// Set stores a key-value pair durably.
// The record is written to the log before updating the in-memory store.
func (db *DB) Set(key string, value []byte) error {
	LogRecord := logstore.Record{
		Type:  logstore.RecordPut,
		Key:   []byte(key),
		Value: value,
	}
	if err := db.log.Append(LogRecord); err != nil {
		return fmt.Errorf("append to log: %w", err)
	}

	db.mem.Set(key, value)

	return nil
}

// Delete removes a key durably.
func (db *DB) Delete(key string) error {
	LogRecord := logstore.Record{
		Type: logstore.RecordDelete,
		Key:  []byte(key),
	}
	if err := db.log.Append(LogRecord); err != nil {
		return fmt.Errorf("append delete to log: %w", err)
	}

	db.mem.Delete(key)

	return nil
}

// Get returns the value for key and whether it was found.
// Reads are served entirely from the in-memory store — no disk access.
func (db *DB) Get(key string) ([]byte, bool) {
	return db.mem.Get(key)
}

// Scan returns a sorted iterator over keys matching prefix.
func (db *DB) Scan(prefix string) iter.Seq2[string, []byte] {
	return db.mem.Scan(prefix)
}

// Close flushes and closes the DB cleanly.
func (db *DB) Close() error {
	db.mem.Stop()

	if err := db.log.Close(); err != nil {
		return fmt.Errorf("close log: %w", err)
	}

	return nil
}
