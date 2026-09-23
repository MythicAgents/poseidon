package files

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadUnknownSizeFile(t *testing.T) {
	expected := []byte("generated\x00content")

	actual, err := readUnknownSizeFile(bytes.NewReader(expected))
	if err != nil {
		t.Fatalf("readUnknownSizeFile returned an error: %v", err)
	}
	if !bytes.Equal(actual, expected) {
		t.Fatalf("readUnknownSizeFile returned %q, want %q", actual, expected)
	}
}

func TestReadUnknownSizeFileRejectsDataOverLimit(t *testing.T) {
	data := bytes.NewReader(make([]byte, maxUnknownFileSize+1))

	_, err := readUnknownSizeFile(data)
	if err == nil {
		t.Fatal("readUnknownSizeFile accepted data over the limit")
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("readUnknownSizeFile returned unexpected error: %v", err)
	}
}
