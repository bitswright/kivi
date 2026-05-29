package logstore

import (
	"os"
	"path/filepath"
	"testing"
)

// tempLog creates a temporary log file for testing.
// Automatically deleted when the test ends.
func tempLog(t *testing.T, mode SyncMode) *LogStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.log")
	l, err := Open(path, mode)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	t.Cleanup(func() { l.Close() })
	return l
}

func TestAppendAndReplay(t *testing.T) {
	l := tempLog(t, SyncNone)

	records := []Record{
		{Type: RecordPut, Key: []byte("foo"), Value: []byte("bar")},
		{Type: RecordPut, Key: []byte("baz"), Value: []byte("qux")},
		{Type: RecordDelete, Key: []byte("foo")},
	}

	for _, rec := range records {
		if err := l.Append(rec); err != nil {
			t.Fatalf("append: %v", err)
		}
	}

	var replayed []Record
	if err := l.Replay(func(rec Record) error {
		replayed = append(replayed, rec)
		return nil
	}); err != nil {
		t.Fatalf("replay: %v", err)
	}

	if len(replayed) != len(records) {
		t.Fatalf("expected %d records, got %d", len(records), len(replayed))
	}

	for i, rec := range replayed {
		if rec.Type != records[i].Type {
			t.Errorf("record %d: type mismatch", i)
		}
		if string(rec.Key) != string(records[i].Key) {
			t.Errorf("record %d: key mismatch", i)
		}
		if string(rec.Value) != string(records[i].Value) {
			t.Errorf("record %d: value mismatch", i)
		}
	}
}

func TestCorruptedRecordSkipped(t *testing.T) {
	path := filepath.Join(t.TempDir() + "test.log")
	l, err := Open(path, SyncNone)
	if err != nil {
		t.Fatal(err)
	}

	recordType := RecordPut
	key := "key"
	value := "value"

	// write a valid record and close file
	l.Append(Record{Type: recordType, Key: []byte(key), Value: []byte(value)})
	l.Close()

	// write garbage value
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatal("open file to append garbage value: %w", err)
	}
	f.Write([]byte("garbage value"))
	f.Close()

	// Open file again using LogStore and Replay
	l2, err := Open(path, SyncNone)
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	err = l2.Replay(func(rec Record) error {
		count++
		return nil
	})

	if count != 1 {
		t.Fatalf("expected 1 valid record, got %d", count)
	}

	if err == nil {
		t.Fatalf("expected error from corrupted record, got nil")
	}
}

func TestReplayEmptyLog(t *testing.T) {
	l := tempLog(t, SyncNone)

	var count int
	err := l.Replay(func(rec Record) error {
		count++
		return nil
	})

	if count != 0 {
		t.Fatalf("expected 0 valid record, got %d", count)
	}

	if err != nil {
		t.Fatalf("unexpected error on empty log: %v", err)
	}
}

func TestDeleteRecordRoundTrip(t *testing.T) {
	l := tempLog(t, SyncNone)

	l.Append(Record{Type: RecordPut, Key: []byte("x"), Value: []byte("hello")})
	l.Append(Record{Type: RecordDelete, Key: []byte("x")})

	types := make([]RecordType, 0)
	l.Replay(func(rec Record) error {
		types = append(types, rec.Type)
		return nil
	})

	if types[0] != RecordPut || types[1] != RecordDelete {
		t.Fatalf("unexpected record types: %v", types)
	}
}

func TestPartialWriteAtEOF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	l, _ := Open(path, SyncNone)

	l.Append(Record{Type: RecordPut, Key: []byte("key"), Value: []byte("val")})
	l.Close()

	// simulate partial write: truncate file to cut the last record in half
	info, _ := os.Stat(path)
	os.Truncate(path, info.Size()-5)

	l2, _ := Open(path, SyncNone)
	defer l2.Close()

	var count int
	err := l2.Replay(func(rec Record) error {
		count++
		return nil
	})

	// partial record should cause an error
	if err == nil {
		t.Fatal("expected error from partial record")
	}
}
