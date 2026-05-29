package logstore

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

type RecordType uint8

const (
	RecordPut        RecordType = 1
	RecordDelete     RecordType = 2
	RecordCheckPoint RecordType = 3
)

// headerSize is the fixed size of every record header in bytes.
// layout: type(1) + keylen(4) + vallen(4) + crc32(4) = 13 bytes
const headerSize = 13

type Record struct {
	Type  RecordType
	Key   []byte
	Value []byte
}

// EncodeRecord writes a single record to w in binary format.
func EncodeRecord(w io.Writer, rec Record) error {
	keyLen := uint32(len(rec.Key))
	valLen := uint32(len(rec.Value))

	// compute checksum over: type + keyLen + valLen + key + val
	h := crc32.NewIEEE()
	h.Write([]byte{byte(rec.Type)})
	// using little-endian, because x86 and ARM processors are natively little-endian
	// hence, no byte swapping while reading on these architectures — slightly faster.
	binary.Write(h, binary.LittleEndian, keyLen)
	binary.Write(h, binary.LittleEndian, valLen)
	h.Write(rec.Key)
	h.Write(rec.Value)
	checksum := h.Sum32()

	// write header
	header := make([]byte, headerSize)
	header[0] = byte(rec.Type)
	binary.LittleEndian.PutUint32(header[1:5], keyLen)
	binary.LittleEndian.PutUint32(header[5:9], valLen)
	binary.LittleEndian.PutUint32(header[9:13], checksum)

	if _, err := w.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// write key
	if _, err := w.Write(rec.Key); err != nil {
		return fmt.Errorf("write key: %w", err)
	}

	// write val
	if len(rec.Value) > 0 {
		if _, err := w.Write(rec.Value); err != nil {
			return fmt.Errorf("write value: %w", err)
		}
	}

	return nil
}

// DecodeRecord reads a single record from r.
// Returns io.EOF if there are no more records.
// Returns a descriptive error if the record is corrupted.
func DecodeRecord(r io.Reader) (Record, error) {
	// read fixed-sized header
	header := make([]byte, headerSize)
	if _, err := io.ReadFull(r, header); err != nil {
		// io.EOF here means clean end of file — no more records
		return Record{}, err
	}
	recType := RecordType(header[0])
	keyLen := binary.LittleEndian.Uint32(header[1:5])
	valLen := binary.LittleEndian.Uint32(header[5:9])
	storedChecksum := binary.LittleEndian.Uint32(header[9:13])

	// read key
	key := make([]byte, keyLen)
	if _, err := io.ReadFull(r, key); err != nil {
		return Record{}, fmt.Errorf("read key: %w", err)
	}

	// read val
	val := make([]byte, valLen)
	if valLen > 0 {
		if _, err := io.ReadFull(r, val); err != nil {
			return Record{}, fmt.Errorf("read val: %w", err)
		}
	}

	// verify checksum
	h := crc32.NewIEEE()
	h.Write([]byte{byte(recType)})
	binary.Write(h, binary.LittleEndian, keyLen)
	binary.Write(h, binary.LittleEndian, valLen)
	h.Write(key)
	h.Write(val)
	computedChecksum := h.Sum32()

	if computedChecksum != storedChecksum {
		return Record{}, fmt.Errorf("checksum mismatch: stored %08x, computed %08x", storedChecksum, computedChecksum)
	}

	return Record{
		Type:  recType,
		Key:   key,
		Value: val,
	}, nil
}
