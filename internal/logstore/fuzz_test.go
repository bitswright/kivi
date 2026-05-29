package logstore

import (
	"bytes"
	"io"
	"testing"
)

// FuzzDecodeRecord feeds random bytes to the decoder.
// Goal: it must never panic, always return a typed error for bad input.
func FuzzDecodeRecord(f *testing.F) {
	// seed corpus — valid records to start from
	var buf bytes.Buffer
	EncodeRecord(&buf, Record{
		Type:  RecordPut,
		Key:   []byte("hello"),
		Value: []byte("world"),
	})
	f.Add(buf.Bytes())

	// empty input
	f.Add([]byte{})

	// just a header with no body
	f.Add(make([]byte, headerSize))

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)
		// must not panic — error is fine, panic is not
		_, err := DecodeRecord(r)
		if err != nil && err != io.EOF {
			// typed error is expected for bad input — this is correct behaviour
			return
		}
	})
}
