package ipc

import (
	"encoding/binary"
	"errors"
	"io"
)

// MaxFrameSize 是单个 IPC 帧允许的最大 payload 字节数。
const MaxFrameSize = 64 * 1024 * 1024

var (
	// ErrFrameTooLarge 表示帧长度超过限制。
	ErrFrameTooLarge = errors.New("ipc: frame exceeds size limit")
	// ErrMalformedFrame 表示帧写入端违反 io.Writer 契约。
	ErrMalformedFrame = errors.New("ipc: malformed frame")
)

func writeFrame(w io.Writer, payload []byte) error {
	if len(payload) > MaxFrameSize {
		return ErrFrameTooLarge
	}

	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	if err := writeAll(w, header[:]); err != nil {
		return err
	}
	return writeAll(w, payload)
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if n < 0 || n > len(data) {
			return ErrMalformedFrame
		}
		data = data[n:]
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}

func readFrame(r io.Reader) ([]byte, error) {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}

	size := binary.BigEndian.Uint32(header[:])
	if size > MaxFrameSize {
		return nil, ErrFrameTooLarge
	}

	payload := make([]byte, int(size))
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	return payload, nil
}
