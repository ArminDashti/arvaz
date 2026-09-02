package iperf

import (
	"io"
)

const (
	DefaultBytes int64 = 8 << 20  // 8 MiB
	MinBytes     int64 = 1 << 20  // 1 MiB
	MaxBytes     int64 = 64 << 20 // 64 MiB
	ChunkSize          = 64 << 10 // 64 KiB write buffer
)

// ClampBytes returns n limited to [MinBytes, MaxBytes]. Zero or negative uses DefaultBytes.
func ClampBytes(n int64) int64 {
	if n <= 0 {
		return DefaultBytes
	}
	if n < MinBytes {
		return MinBytes
	}
	if n > MaxBytes {
		return MaxBytes
	}
	return n
}

// WriteZeros writes n zero bytes to w in ChunkSize pieces.
func WriteZeros(w io.Writer, n int64) error {
	buf := make([]byte, ChunkSize)
	remaining := n
	for remaining > 0 {
		chunk := int64(len(buf))
		if chunk > remaining {
			chunk = remaining
		}
		if _, err := w.Write(buf[:chunk]); err != nil {
			return err
		}
		remaining -= chunk
	}
	return nil
}

// DiscardBody copies r to io.Discard and returns bytes read.
func DiscardBody(r io.Reader) (int64, error) {
	return io.Copy(io.Discard, r)
}
