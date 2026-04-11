package wal

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type WAL struct {
	file *os.File
}

func Open(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("could not open WAL file: %w", err)
	}
	return &WAL{file: file}, nil
}

func (w *WAL) Write(e Entry) error {
	buf := encode(e)

	if _, err := w.file.Write(buf); err != nil {
		return fmt.Errorf("could not write entry to WAL file: %w", err)
	}

	// force OS to flush buffer to physical disk
	if err := w.file.Sync(); err != nil {
		return fmt.Errorf("could not sync WAL file: %w", err)
	}

	return nil
}

func (w *WAL) Replay() ([]Entry, error) {
	if _, err := w.file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("could not seek to start of WAL file: %w", err)
	}

	var entries []Entry
	for {
		// read fixed-size header first: op(1) + key_len(4) + val_len(4) = 9 bytes
		header := make([]byte, 9)
		if _, err := io.ReadFull(w.file, header); err != nil {
			if err == io.EOF {
				break // clean end of file
			}
			break // partial header
		}

		keyLen := binary.LittleEndian.Uint32(header[1:5])
		valLen := binary.LittleEndian.Uint32(header[5:9])

		restEntry := make([]byte, keyLen+valLen+4)
		if _, err := io.ReadFull(w.file, restEntry); err != nil {
			break // partial entry
		}

		buf := append(header, restEntry...)
		entry, ok := decode(buf)
		if !ok {
			break // corrupted entry, stop here
		}

		entries = append(entries, entry)
	}
	return entries, nil
}

func (w *WAL) Close() error {
	return w.file.Close()
}
