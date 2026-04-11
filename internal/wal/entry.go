package wal

import (
	"encoding/binary"
	"hash/crc32"
)

type OpType uint8

const (
	OpSet OpType = iota
	OpDelete
)

type Entry struct {
	Op    OpType
	Key   string
	Value string
}

func encode(e Entry) []byte {
	keyBytes := []byte(e.Key)
	valBytes := []byte(e.Value)

	keyLen := uint32(len(keyBytes))
	valLen := uint32(len(valBytes))

	buf := make([]byte, 1+4+4+keyLen+valLen+4)

	buf[0] = byte(e.Op)
	binary.LittleEndian.PutUint32(buf[1:5], keyLen)
	// using LittleEndian (LSB is first byte) to keep coding and decoding consistent for different machines
	binary.LittleEndian.PutUint32(buf[5:9], valLen)
	copy(buf[9:9+keyLen], keyBytes)
	copy(buf[9+keyLen:9+keyLen+valLen], valBytes)

	checksum := crc32.ChecksumIEEE(buf[:9+keyLen+valLen])
	binary.LittleEndian.PutUint32(buf[9+keyLen+valLen:], checksum)

	return buf
}

func decode(buf []byte) (Entry, bool) {
	if len(buf) < 13 {
		return Entry{}, false
	}

	opType := OpType(buf[0])
	keyLen := binary.LittleEndian.Uint32(buf[1:5])
	valLen := binary.LittleEndian.Uint32(buf[5:9])

	expectedLen := int(13 + keyLen + valLen)
	if len(buf) != expectedLen {
		return Entry{}, false
	}

	key := string(buf[9 : 9+keyLen])
	val := string(buf[9+keyLen : 9+keyLen+valLen])

	storedChecksum := binary.LittleEndian.Uint32(buf[9+keyLen+valLen:])
	expectedChecksum := crc32.ChecksumIEEE(buf[:9+keyLen+valLen])

	if storedChecksum != expectedChecksum {
		return Entry{}, false
	}

	return Entry{
		Op:    opType,
		Key:   key,
		Value: val,
	}, true
}
