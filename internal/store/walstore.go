package store

import (
	"fmt"

	"github.com/bitswright/kivi/internal/wal"
)

type WALStore struct {
	MemStore
	wal *wal.WAL
}

func NewWALStore(path string) (*WALStore, error) {
	w, err := wal.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open WAL: %w", err)
	}

	ws := &WALStore{
		MemStore: MemStore{data: make(map[string]string)},
		wal:      w,
	}

	if err := ws.replay(); err != nil {
		return nil, fmt.Errorf("could not replay WAL: %w", err)
	}

	return ws, nil
}

func (ws *WALStore) Set(key, value string) error {
	entry := wal.Entry{
		Op:    wal.OpSet,
		Key:   key,
		Value: value,
	}

	if err := ws.wal.Write(entry); err != nil {
		return err
	}

	return ws.MemStore.Set(key, value)
}

func (ws *WALStore) Delete(key string) error {
	entry := wal.Entry{
		Op:  wal.OpDelete,
		Key: key,
	}

	if err := ws.wal.Write(entry); err != nil {
		return err
	}

	return ws.MemStore.Delete(key)
}

func (ws *WALStore) replay() error {
	entries, err := ws.wal.Replay()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		switch entry.Op {
		case wal.OpSet:
			ws.MemStore.Set(entry.Key, entry.Value)
		case wal.OpDelete:
			ws.MemStore.Delete(entry.Key)
		}
	}

	return nil
}

func (ws *WALStore) Close() error {
	return ws.wal.Close()
}
