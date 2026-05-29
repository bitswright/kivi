package logstore

import (
	"fmt"
	"io"
	"os"
)

// SyncMode controls how aggressively writes are flushed to disk.
type SyncMode int

const (
	// SyncNone — OS decides when to flush. Fastest, least durable.
	// Data may be lost if the machine loses power before OS flushes.
	SyncNone SyncMode = iota

	// SyncData — flush data to disk but not metadata (file size, timestamps).
	// Good balance of performance and durability on Linux.
	SyncData

	// SyncFull — flush data and metadata. Slowest, most durable.
	// Guarantees the record survives a power loss.
	SyncFull
)

// LogStore is an append-only log that persists records to disk.
type LogStore struct {
	file     *os.File
	syncMode SyncMode
}

func Open(path string, mode SyncMode) (*LogStore, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	return &LogStore{
		file:     f,
		syncMode: mode,
	}, nil
}

// Append encodes and appends a record to the log.
func (l *LogStore) Append(rec Record) error {
	if err := EncodeRecord(l.file, rec); err != nil {
		return fmt.Errorf("encode record: %w", err)
	}
	return l.sync()
}

func (l *LogStore) sync() error {
	switch l.syncMode {
	case SyncNone:
		return nil
	case SyncData:
		// Go stdlib exposes only Sync(); fdatasync via syscall if needed
		return l.file.Sync()
	case SyncFull:
		return l.file.Sync()
	}
	return nil
}

// Replay reads the log from the beginning and calls fn for each valid record.
// Stops at the first corrupted record or EOF.
func (l *LogStore) Replay(fn func(Record) error) error {
	// seek to beginning of file
	if _, err := l.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek to start: %w", err)
	}

	for {
		rec, err := DecodeRecord(l.file)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			// corrupted or partial record — log it and stop
			return fmt.Errorf("replay stopped: %w", err)
		}
		if err := fn(rec); err != nil {
			return fmt.Errorf("replay fn: %w", err)
		}
	}
}

// Sync explicitly flushes the log to disk regardless of SyncMode.
func (l *LogStore) Sync() error {
	return l.file.Sync()
}

// Close syncs and closes the log file.
func (l *LogStore) Close() error {
	if err := l.file.Sync(); err != nil {
		return fmt.Errorf("sync on close: %w", err)
	}
	return l.file.Close()
}
