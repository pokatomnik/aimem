package main

import "testing"

// With no LOG_FILE the logger must still be usable: stderr is the default
// destination, and the closer has to be callable so main can defer it
// unconditionally. The test never writes, so it touches no file.
func TestNewLoggerWithoutAFile(t *testing.T) {
	t.Parallel()

	logger, closeLog, err := newLogger(&Config{LogLevel: "info"})
	if err != nil {
		t.Fatalf("newLogger: %v", err)
	}
	defer closeLog()

	if logger == nil {
		t.Fatal("logger is nil")
	}
}
