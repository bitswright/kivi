package store

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type LogStore struct {
	MemStore
	file *os.File
}

func NewLogStore(path string) (*LogStore, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("could not open log file: %w", err)
	}

	ls := &LogStore{
		MemStore: MemStore{data: make(map[string]string)},
		file:     file,
	}

	if err := ls.replay(); err != nil {
		return nil, fmt.Errorf("could not replay log: %w", err)
	}

	return ls, nil
}

func (ls *LogStore) Set(key, value string) error {
	if _, err := fmt.Fprintf(ls.file, "SET %s %s\n", key, value); err != nil {
		return fmt.Errorf("could not write to log: %w", err)
	}
	return ls.MemStore.Set(key, value)
}

func (ls *LogStore) Delete(key string) error {
	if _, err := fmt.Fprintf(ls.file, "DEL %s\n", key); err != nil {
		return fmt.Errorf("could not write to log: %w", err)
	}
	return ls.MemStore.Delete(key)
}

func (ls *LogStore) replay() error {
	if _, err := ls.file.Seek(0, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(ls.file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)

		switch parts[0] {
		case "SET":
			if len(parts) != 3 {
				// skip malformed entry
				continue
			}
			ls.MemStore.Set(parts[1], parts[2])
		case "DEL":
			if len(parts) != 2 {
				// skip malformed entry
				continue
			}
			ls.MemStore.Delete(parts[1])
		}
	}

	return scanner.Err()
}

func (ls *LogStore) Close() error {
	return ls.file.Close()
}
