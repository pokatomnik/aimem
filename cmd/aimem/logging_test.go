package main

import (
	"io"
	"os"
	"testing"
)

func TestOpenLogFileWithoutPathUsesStderr(t *testing.T) {
	t.Parallel()

	file, closeLogFile, err := openLogFile("")
	if err != nil {
		t.Fatalf("openLogFile: %v", err)
	}
	if file != os.Stderr {
		t.Fatalf("openLogFile() = %p, want stderr %p", file, os.Stderr)
	}
	if closeLogFile == nil {
		t.Fatal("openLogFile() returned a nil close callback")
	}
	closeLogFile()
}

func TestNewLogger(t *testing.T) {
	t.Parallel()

	if logger := newLogger(io.Discard, "info"); logger == nil {
		t.Fatal("newLogger returned nil")
	}
}
