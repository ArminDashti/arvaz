package iperf

import (
	"bytes"
	"io"
	"testing"
)

func TestClampBytes(t *testing.T) {
	tests := []struct {
		in   int64
		want int64
	}{
		{0, DefaultBytes},
		{-1, DefaultBytes},
		{MinBytes - 1, MinBytes},
		{MinBytes, MinBytes},
		{DefaultBytes, DefaultBytes},
		{MaxBytes, MaxBytes},
		{MaxBytes + 1, MaxBytes},
	}
	for _, tc := range tests {
		if got := ClampBytes(tc.in); got != tc.want {
			t.Fatalf("ClampBytes(%d)=%d want %d", tc.in, got, tc.want)
		}
	}
}

func TestWriteZeros(t *testing.T) {
	var buf bytes.Buffer
	n := MinBytes
	if err := WriteZeros(&buf, n); err != nil {
		t.Fatal(err)
	}
	if int64(buf.Len()) != n {
		t.Fatalf("len=%d want %d", buf.Len(), n)
	}
}

func TestDiscardBody(t *testing.T) {
	payload := bytes.Repeat([]byte{7}, int(MinBytes))
	got, err := DiscardBody(bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if got != MinBytes {
		t.Fatalf("got %d want %d", got, MinBytes)
	}
	n, err := io.Copy(io.Discard, bytes.NewReader(nil))
	if err != nil || n != 0 {
		t.Fatalf("empty body n=%d err=%v", n, err)
	}
}
