package ipc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

type limitedReader struct {
	reader *bytes.Reader
	limit  int
}

func (r limitedReader) Read(p []byte) (int, error) {
	if len(p) > r.limit {
		p = p[:r.limit]
	}
	return r.reader.Read(p)
}

type limitedWriter struct {
	bytes.Buffer
	limit int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if len(p) > w.limit {
		p = p[:w.limit]
	}
	return w.Buffer.Write(p)
}

type zeroWriter struct{}

func (zeroWriter) Write([]byte) (int, error) { return 0, nil }

func TestFrameRoundTripWithShortReadsAndWrites(t *testing.T) {
	var output limitedWriter
	output.limit = 2
	want := []byte(`{"cmd":"Get"}`)

	if err := writeFrame(&output, want); err != nil {
		t.Fatalf("writeFrame() error = %v", err)
	}

	got, err := readFrame(limitedReader{reader: bytes.NewReader(output.Bytes()), limit: 1})
	if err != nil {
		t.Fatalf("readFrame() error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("readFrame() = %q, want %q", got, want)
	}
}

func TestReadFrameAllowsEmptyPayload(t *testing.T) {
	var frame [4]byte
	got, err := readFrame(bytes.NewReader(frame[:]))
	if err != nil {
		t.Fatalf("readFrame() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("readFrame() length = %d, want 0", len(got))
	}
}

func TestReadFrameRejectsOversizedLengthBeforeAllocation(t *testing.T) {
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(MaxFrameSize+1))

	_, err := readFrame(bytes.NewReader(header[:]))
	if !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("readFrame() error = %v, want %v", err, ErrFrameTooLarge)
	}
}

func TestWriteFrameRejectsZeroProgressWriter(t *testing.T) {
	err := writeFrame(zeroWriter{}, []byte("x"))
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("writeFrame() error = %v, want %v", err, io.ErrShortWrite)
	}
}

func TestReadFrameRejectsTruncatedPayload(t *testing.T) {
	frame := []byte{0, 0, 0, 2, 'x'}
	_, err := readFrame(bytes.NewReader(frame))
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("readFrame() error = %v, want %v", err, io.ErrUnexpectedEOF)
	}
}
