package logstore

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func benchmarkAppend(b *testing.B, mode SyncMode) {
	path := filepath.Join(b.TempDir(), "bench.log")
	l, err := Open(path, mode)
	if err != nil {
		b.Fatalf("open logstore: %v", err)
	}
	defer l.Close()
	defer os.Remove(path)

	rec := Record{
		Type:  RecordPut,
		Key:   []byte("bench-key"),
		Value: []byte("bench-value"),
	}

	b.ResetTimer()

	for range b.N {
		if err := l.Append(rec); err != nil {
			b.Fatalf("append: %v", err)
		}
	}
}

func BenchmarkAppendSyncNone(b *testing.B) {
	benchmarkAppend(b, SyncNone)
}

func BenchmarkAppendSyncData(b *testing.B) {
	benchmarkAppend(b, SyncData)
}

func BenchmarkAppendSyncFull(b *testing.B) {
	benchmarkAppend(b, SyncFull)
}

func BenchmarkReplay(b *testing.B) {
	path := filepath.Join(b.TempDir(), "bench.log")
	l, err := Open(path, SyncNone)
	if err != nil {
		b.Fatalf("open logstore: %v", err)
	}

	// pre-populate with 10000 records and close the file
	for i := range 10000 {
		l.Append(Record{
			Type:  RecordPut,
			Key:   []byte(fmt.Sprintf("key-%d", i)),
			Value: []byte("value"),
		})
	}
	l.Close()

	b.ResetTimer()

	for range b.N {
		l2, _ := Open(path, SyncNone)
		l2.Replay(func(rec Record) error { return nil })
		l2.Close()
	}
}
