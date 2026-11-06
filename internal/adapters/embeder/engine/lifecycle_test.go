package engine

import (
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

// closingFS hands out files that refuse every read after Close, which fstest.MapFS
// and embed.FS do not: an embed file's Close is a no-op and a MapFS file owns no
// descriptor, so neither punishes a loader that closes too early. Only os.DirFS
// does, and a hermetic test may not use it.
//
// This exists because the first draft of safetensors.go closed the checkpoint in
// a defer inside openSafetensors and then read tensors from it. The golden test
// stayed green — every tensor read succeeded, because the file was never really
// closed — and the bug surfaced only against a real directory. A loader that
// keeps the file alive must fail these tests if it ever regresses.
type closingFS struct {
	inner fs.FS
}

func (c closingFS) Open(name string) (fs.File, error) {
	f, err := c.inner.Open(name)
	if err != nil {
		return nil, err
	}
	return &closingFile{File: f}, nil
}

type closingFile struct {
	fs.File
	closed bool
}

func (f *closingFile) Read(p []byte) (int, error) {
	if f.closed {
		return 0, fs.ErrClosed
	}
	return f.File.Read(p)
}

func (f *closingFile) ReadAt(p []byte, off int64) (int, error) {
	if f.closed {
		return 0, fs.ErrClosed
	}
	return f.File.(io.ReaderAt).ReadAt(p, off)
}

func (f *closingFile) Stat() (fs.FileInfo, error) {
	if f.closed {
		return nil, fs.ErrClosed
	}
	return f.File.Stat()
}

// Close is idempotent; a loader may reasonably close twice.
func (f *closingFile) Close() error {
	if f.closed {
		return nil
	}
	f.closed = true
	return f.File.Close()
}

func goldenMapFS(t *testing.T) fstest.MapFS {
	t.Helper()
	raw, err := goldenFS.ReadFile(goldenDir + "/model.safetensors")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return fstest.MapFS{"model.safetensors": &fstest.MapFile{Data: raw}}
}

// TestCheckpointFileOutlivesEveryTensorRead pins the lifecycle: every tensor is
// read while the checkpoint file is still open. A loader that closes it inside
// openSafetensors fails on the first tensor.
func TestCheckpointFileOutlivesEveryTensorRead(t *testing.T) {
	st, err := openSafetensors(closingFS{inner: goldenMapFS(t)}, "model.safetensors")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if len(st.tensors) == 0 {
		t.Fatal("checkpoint has no tensors")
	}
	for name := range st.tensors {
		if _, err := st.TensorF32(name); err != nil {
			t.Fatalf("read %s while open: %v", name, err)
		}
	}
	if err := st.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	// After Close the handle is gone, and the type must say so rather than hand
	// out bytes; upstream's SafetensorsFile has the same guard.
	for name := range st.tensors {
		if _, err := st.TensorF32(name); err == nil {
			t.Fatalf("TensorF32(%q) succeeded after Close", name)
		}
		break
	}
}

// TestClosedFileIsNotReopened is the direct regression on the fake itself: a read
// after Close must fail, so a loader that closes early cannot pass silently.
func TestClosedFileIsNotReopened(t *testing.T) {
	f, err := closingFS{inner: goldenMapFS(t)}.Open("model.safetensors")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.(io.ReaderAt).ReadAt(make([]byte, 16), 0); err != nil {
		t.Fatalf("read while open: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := f.(io.ReaderAt).ReadAt(make([]byte, 16), 0); err == nil {
		t.Fatal("ReadAt succeeded on a closed file: the fake does not model a closed descriptor")
	}
}

// TestTensorSurvivesClose guards the other half of the contract: the returned
// slice owns its memory, so the caller never has to keep the file alive. That is
// what lets LoadEncoder close the checkpoint and still return a usable encoder.
func TestTensorSurvivesClose(t *testing.T) {
	mfs := goldenMapFS(t)

	st, err := openSafetensors(mfs, "model.safetensors")
	if err != nil {
		t.Fatal(err)
	}
	want, err := st.TensorF32("language_model.norm.weight")
	if err != nil {
		t.Fatal(err)
	}
	if len(want) == 0 {
		t.Fatal("empty tensor")
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	// The values read before Close must equal a fresh read: nothing aliased the
	// file, so closing changed nothing.
	st2, err := openSafetensors(mfs, "model.safetensors")
	if err != nil {
		t.Fatal(err)
	}
	again, err := st2.TensorF32("language_model.norm.weight")
	if err != nil {
		t.Fatal(err)
	}
	for i := range want {
		if want[i] != again[i] {
			t.Fatalf("value %d changed after Close: %v != %v", i, want[i], again[i])
		}
	}
}
