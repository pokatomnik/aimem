package engine

import (
	"encoding/binary"
	"encoding/json"
	"io/fs"
	"math"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

// The tests in this file build safetensors blobs in memory. They are about this
// package's own rules — which headers it refuses and how it widens dtypes — not
// about the checkpoint, so a few hundred bytes of synthetic data is the right
// fixture: hermetic, fast, and every case states exactly what it exercises.

type spec struct {
	dtype string
	shape []int
	raw   []byte
}

// encodeF32 is little-endian float32, as the format writes it.
func encodeF32(vs ...float32) []byte {
	out := make([]byte, len(vs)*4)
	for i, v := range vs {
		binary.LittleEndian.PutUint32(out[i*4:], math.Float32bits(v))
	}
	return out
}

// encodeBF16 truncates a float32 to bfloat16. The values used by the tests are
// exactly representable in bf16, so the widening round-trips.
func encodeBF16(vs ...float32) []byte {
	out := make([]byte, len(vs)*2)
	for i, v := range vs {
		binary.LittleEndian.PutUint16(out[i*2:], uint16(math.Float32bits(v)>>16))
	}
	return out
}

// buildSafetensors assembles a header and payload. Names are sorted so the
// offsets are deterministic.
func buildSafetensors(t *testing.T, specs map[string]spec) []byte {
	t.Helper()
	type entry struct {
		DType       string   `json:"dtype"`
		Shape       []int    `json:"shape"`
		DataOffsets [2]int64 `json:"data_offsets"`
	}
	names := make([]string, 0, len(specs))
	for n := range specs {
		names = append(names, n)
	}
	sort.Strings(names)

	header := make(map[string]entry, len(specs))
	var payload []byte
	for _, n := range names {
		s := specs[n]
		lo := int64(len(payload))
		payload = append(payload, s.raw...)
		header[n] = entry{DType: s.dtype, Shape: s.shape, DataOffsets: [2]int64{lo, int64(len(payload))}}
	}
	rawHeader, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}

	out := make([]byte, 8+len(rawHeader)+len(payload))
	binary.LittleEndian.PutUint64(out, uint64(len(rawHeader)))
	copy(out[8:], rawHeader)
	copy(out[8+len(rawHeader):], payload)
	return out
}

func safetensorsFS(blob []byte) fstest.MapFS {
	return fstest.MapFS{"model.safetensors": &fstest.MapFile{Data: blob}}
}

// TestTensorF32WidensBF16 pins the widening arithmetic: bf16 is a float32 whose
// low 16 mantissa bits are zero, so the values chosen here survive the round trip
// exactly and any change to the shift fails.
func TestTensorF32WidensBF16(t *testing.T) {
	want := []float32{1, -2.5, 0.75, 0, -0.125}
	blob := buildSafetensors(t, map[string]spec{
		"w": {dtype: "BF16", shape: []int{len(want)}, raw: encodeBF16(want...)},
	})

	st, err := openSafetensors(safetensorsFS(blob), "model.safetensors")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	got, err := st.TensorF32("w", len(want))
	if err != nil {
		t.Fatalf("TensorF32: %v", err)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("value %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

// TestTensorF32ReadsF32 covers the other dtype the real checkpoint could use.
func TestTensorF32ReadsF32(t *testing.T) {
	want := []float32{3.25, -0.5, 1e10}
	blob := buildSafetensors(t, map[string]spec{
		"w": {dtype: "F32", shape: []int{len(want)}, raw: encodeF32(want...)},
	})

	st, err := openSafetensors(safetensorsFS(blob), "model.safetensors")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	got, err := st.TensorF32("w")
	if err != nil {
		t.Fatalf("TensorF32: %v", err)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("value %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

// TestDecodeCrossesChunkBoundary exercises the chunked reader: a tensor larger
// than decodeChunk is filled across several ReadAt calls, and every element must
// land in its own slot.
func TestDecodeCrossesChunkBoundary(t *testing.T) {
	n := decodeChunk*2 + 7
	want := make([]float32, n)
	for i := range want {
		want[i] = float32(i % 1000)
	}
	blob := buildSafetensors(t, map[string]spec{
		"w": {dtype: "F32", shape: []int{n}, raw: encodeF32(want...)},
	})

	st, err := openSafetensors(safetensorsFS(blob), "model.safetensors")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	got, err := st.TensorF32("w", n)
	if err != nil {
		t.Fatalf("TensorF32: %v", err)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("value %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

// TestRejectsShapePayloadMismatch is the validation upstream calls H2: a header
// may not pair a shape with a byte range that cannot hold it, or the caller
// indexing by shape panics at inference.
func TestRejectsShapePayloadMismatch(t *testing.T) {
	blob := buildSafetensors(t, map[string]spec{
		"w": {dtype: "BF16", shape: []int{4}, raw: encodeBF16(1, 2)},
	})
	if _, err := openSafetensors(safetensorsFS(blob), "model.safetensors"); err == nil {
		t.Fatal("accepted a shape that does not match its byte range")
	}
}

func TestRejectsNegativeDim(t *testing.T) {
	blob := buildSafetensors(t, map[string]spec{
		"w": {dtype: "F32", shape: []int{-1}, raw: nil},
	})
	if _, err := openSafetensors(safetensorsFS(blob), "model.safetensors"); err == nil {
		t.Fatal("accepted a negative dimension")
	}
}

// TestRejectsOffsetsBeyondPayload: the byte range must lie inside the payload.
func TestRejectsOffsetsBeyondPayload(t *testing.T) {
	type entry struct {
		DType       string   `json:"dtype"`
		Shape       []int    `json:"shape"`
		DataOffsets [2]int64 `json:"data_offsets"`
	}
	header, err := json.Marshal(map[string]entry{
		"w": {DType: "F32", Shape: []int{1}, DataOffsets: [2]int64{0, 100}},
	})
	if err != nil {
		t.Fatal(err)
	}
	blob := make([]byte, 8+len(header)+4)
	binary.LittleEndian.PutUint64(blob, uint64(len(header)))
	copy(blob[8:], header)

	if _, err := openSafetensors(safetensorsFS(blob), "model.safetensors"); err == nil {
		t.Fatal("accepted data_offsets past the payload")
	}
}

// TestRejectsTruncatedHeader: the length prefix may not exceed the file, or the
// header slice below panics.
func TestRejectsTruncatedHeader(t *testing.T) {
	blob := make([]byte, 16)
	binary.LittleEndian.PutUint64(blob, math.MaxUint64)
	if _, err := openSafetensors(safetensorsFS(blob), "model.safetensors"); err == nil {
		t.Fatal("accepted a header length larger than the file")
	}

	short := []byte{1, 2, 3}
	if _, err := openSafetensors(safetensorsFS(short), "model.safetensors"); err == nil {
		t.Fatal("accepted a file shorter than its length prefix")
	}
}

// TestUnsupportedDType: a dtype the reader cannot widen parses (it may simply
// never be read) but must be refused at the point of use.
func TestUnsupportedDType(t *testing.T) {
	blob := buildSafetensors(t, map[string]spec{
		"w": {dtype: "F8_E4M3", shape: []int{2}, raw: []byte{0, 0}},
	})
	st, err := openSafetensors(safetensorsFS(blob), "model.safetensors")
	if err != nil {
		t.Fatalf("open should not reject an unknown dtype: %v", err)
	}
	if _, err := st.TensorF32("w"); err == nil {
		t.Fatal("TensorF32 accepted an unsupported dtype")
	}
}

// TestRejectsNonSeekableFS: the loader needs random access, so a file that only
// offers sequential Read must be rejected by name rather than misbehaving.
func TestRejectsNonSeekableFS(t *testing.T) {
	blob := buildSafetensors(t, map[string]spec{
		"w": {dtype: "F32", shape: []int{1}, raw: encodeF32(1)},
	})
	// Embedding fs.File in a struct promotes only Read, Stat and Close: the
	// concrete file's ReadAt is not reachable through the interface.
	seq := sequentialFS{inner: safetensorsFS(blob)}
	if _, err := openSafetensors(seq, "model.safetensors"); err == nil {
		t.Fatal("accepted a file without io.ReaderAt")
	}
}

type sequentialFS struct{ inner fstest.MapFS }

func (s sequentialFS) Open(name string) (fs.File, error) {
	f, err := s.inner.Open(name)
	if err != nil {
		return nil, err
	}
	return struct{ fs.File }{f}, nil
}

// TestMissingTensor names the tensor in the error, because a renamed tensor in a
// new checkpoint is the most likely upstream change.
func TestMissingTensor(t *testing.T) {
	blob := buildSafetensors(t, map[string]spec{
		"w": {dtype: "F32", shape: []int{1}, raw: encodeF32(1)},
	})
	st, err := openSafetensors(safetensorsFS(blob), "model.safetensors")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_, err = st.TensorF32("language_model.norm.weight")
	if err == nil {
		t.Fatal("expected an error for a missing tensor")
	}
	if want := "language_model.norm.weight"; !strings.Contains(err.Error(), want) {
		t.Fatalf("error %q does not name %q", err, want)
	}
}

// TestHalfToF32 covers the F16 expansion without needing an F16 checkpoint.
func TestHalfToF32(t *testing.T) {
	for _, tc := range []struct {
		bits uint16
		want float32
	}{
		{0x0000, 0},
		{0x3C00, 1},
		{0xBC00, -1},
		{0x3555, 0.33325195}, // 1/3, rounded to half
		{0x7C00, float32(math.Inf(1))},
		{0xFC00, float32(math.Inf(-1))},
	} {
		if got := halfToF32(tc.bits); got != tc.want {
			t.Errorf("halfToF32(0x%04X) = %v, want %v", tc.bits, got, tc.want)
		}
	}
	if got := halfToF32(0x0001); got <= 0 || math.IsInf(float64(got), 0) {
		t.Errorf("subnormal half decoded to %v", got)
	}
	if got := halfToF32(0x7E00); !math.IsNaN(float64(got)) {
		t.Errorf("NaN half decoded to %v", got)
	}
}
