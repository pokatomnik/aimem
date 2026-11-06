package engine

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
)

// This file replaces aikit's embed.OpenSafetensorsMmap/OpenSafetensorsFromFS
// with a reader that keeps the checkpoint's bytes where they already are: inside
// the binary, reachable through fs.FS, without copying the 1.49 GB file into the
// heap and without a memory mapping that would need a real path.
//
// The header parsing and the validation rules follow aikit v1.76.0
// (embed/safetensors.go, MIT, Copyright (c) 2026 Francis Townsend-Merino): a
// truncated header, offsets outside the payload, a negative dimension, a shape
// that overflows, or a shape whose element count and dtype width disagree with
// the declared byte range are all rejected, so a corrupt header fails at load
// instead of panicking at inference. The dtype widening is aikit's arithmetic,
// verbatim: BF16 is a float32 with the low 16 mantissa bits zeroed, and F16 is
// decoded through the half-to-float expansion below.
//
// What differs is the read path: aikit's heap variant is fs.ReadFile followed by
// zero-copy views into that one big slice, while this reads tensor by tensor
// through io.ReaderAt in bounded chunks (decodeChunk elements at a time), so
// peak memory is the float32 weights plus one chunk, never the whole file.

// dtypeSize is the byte width of the safetensors dtypes this package can size.
// Unknown dtypes are not rejected at parse time (a tensor that is never read
// cannot hurt), but TensorF32 refuses them.
func dtypeSize(dtype string) (int, bool) {
	switch dtype {
	case "F64", "I64":
		return 8, true
	case "F32", "I32":
		return 4, true
	case "BF16", "F16":
		return 2, true
	default:
		return 0, false
	}
}

// headerTensor is one entry of a safetensors header.
type headerTensor struct {
	DType       string   `json:"dtype"`
	Shape       []int    `json:"shape"`
	DataOffsets [2]int64 `json:"data_offsets"`
}

// tensor is a header entry resolved against the payload.
type tensor struct {
	dtype string
	shape []int
	off   int64 // from the payload start
	size  int64 // bytes
}

// safetensorsFile is a checkpoint reachable through io.ReaderAt. Tensor data is
// read on demand; nothing is mapped and nothing is copied up front.
type safetensorsFile struct {
	f       fs.File
	r       io.ReaderAt
	base    int64 // payload start: 8 + header length
	payload int64 // payload length in bytes
	tensors map[string]tensor
	closed  bool
}

// decodeChunk bounds the transient buffer a tensor read uses. 64Ki elements is
// 128 KB per BF16/F16 chunk and 256 KB per F32 chunk: large enough to keep
// ReadAt calls few, small enough to stay out of the way of the weights.
const decodeChunk = 1 << 16

// openSafetensors opens fsys/name and parses its header. The file stays open
// until Close, because tensor data is read lazily; callers hold the *safetensorsFile
// for as long as they read tensors from it.
func openSafetensors(fsys fs.FS, name string) (*safetensorsFile, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}

	r, ok := f.(io.ReaderAt)
	if !ok {
		_ = f.Close()
		return nil, fmt.Errorf("safetensors: %s: %T does not support random access", name, f)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("safetensors: %s: %w", name, err)
	}
	total := info.Size()
	if total < 8 {
		_ = f.Close()
		return nil, fmt.Errorf("safetensors: %s: file too short for a header length prefix (%d bytes)", name, total)
	}

	var prefix [8]byte
	if _, err := readFullAt(r, prefix[:], 0); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("safetensors: %s: %w", name, err)
	}
	// Little-endian, as the format writes it. Compare before adding 8: 8+headerLen
	// wraps uint64 for a hostile length near 2^64, which would pass the bound and
	// then panic on the slice below. total >= 8 is established above.
	headerLen := binary.LittleEndian.Uint64(prefix[:])
	if headerLen > uint64(total)-8 {
		_ = f.Close()
		return nil, fmt.Errorf("safetensors: %s: file truncated (header claims %d bytes, only %d available)", name, headerLen, total-8)
	}
	header := make([]byte, headerLen)
	if _, err := readFullAt(r, header, 8); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("safetensors: %s: %w", name, err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(header, &raw); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("safetensors: %s: parse header: %w", name, err)
	}

	out := &safetensorsFile{
		f:       f,
		r:       r,
		base:    8 + int64(headerLen),
		payload: total - 8 - int64(headerLen),
		tensors: make(map[string]tensor, len(raw)),
	}
	for tensorName, rawJSON := range raw {
		if tensorName == "__metadata__" {
			continue
		}
		var t headerTensor
		if err := json.Unmarshal(rawJSON, &t); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("safetensors: parse tensor %q: %w", tensorName, err)
		}
		lo, hi := t.DataOffsets[0], t.DataOffsets[1]
		if lo < 0 || hi > out.payload || lo > hi {
			_ = f.Close()
			return nil, fmt.Errorf("safetensors: tensor %q has invalid offsets %v (payload size %d)", tensorName, t.DataOffsets, out.payload)
		}
		// The offset range alone does not stop a hostile header from pairing a
		// huge shape with a one-element byte range: that parses, then panics at
		// inference when the caller indexes by shape. Cross-validate the shape
		// and dtype against the declared byte range, overflow-safe.
		if err := checkShape(tensorName, t, hi-lo, out.payload); err != nil {
			_ = f.Close()
			return nil, err
		}
		out.tensors[tensorName] = tensor{dtype: t.DType, shape: t.Shape, off: lo, size: hi - lo}
	}
	return out, nil
}

// checkShape rejects a tensor whose shape and dtype width do not add up to its
// declared byte range.
func checkShape(name string, t headerTensor, bytes, payload int64) error {
	sz, known := dtypeSize(t.DType)
	if !known {
		return nil // unknown dtypes are refused by TensorF32, not here
	}
	n := int64(1)
	for _, dim := range t.Shape {
		d := int64(dim)
		if d < 0 {
			return fmt.Errorf("safetensors: tensor %q has negative dim in shape %v", name, t.Shape)
		}
		// Bound the running product before multiplying, so n*d cannot wrap.
		if d != 0 && n > payload/d {
			return fmt.Errorf("safetensors: tensor %q shape %v exceeds the payload", name, t.Shape)
		}
		n *= d
	}
	if n > math.MaxInt64/int64(sz) || n*int64(sz) != bytes {
		return fmt.Errorf("safetensors: tensor %q shape %v x %s (%dB/elem) = %d bytes, but byte range is %d",
			name, t.Shape, t.DType, sz, n*int64(sz), bytes)
	}
	return nil
}

// TensorF32 returns the named tensor widened to float32 in freshly allocated
// memory, shape-checked against want when want is non-empty. The returned slice
// never aliases the file, so it stays valid after Close.
func (f *safetensorsFile) TensorF32(name string, want ...int) ([]float32, error) {
	if f.closed {
		return nil, fmt.Errorf("safetensors: TensorF32(%q) called after Close", name)
	}
	t, ok := f.tensors[name]
	if !ok {
		return nil, fmt.Errorf("safetensors: tensor %q not found", name)
	}
	if len(want) > 0 && !shapeEqual(t.shape, want) {
		return nil, fmt.Errorf("safetensors: tensor %q shape %v != want %v", name, t.shape, want)
	}

	n := 1
	for _, d := range t.shape {
		n *= d
	}
	dst := make([]float32, n)

	switch t.dtype {
	case "BF16":
		err := f.decode(t.off, dst, 2, func(b []byte, i int) float32 {
			// Little-endian uint16 into the high 16 bits of the float32 pattern.
			return math.Float32frombits(uint32(binary.LittleEndian.Uint16(b[i*2:])) << 16)
		})
		if err != nil {
			return nil, err
		}
	case "F16":
		err := f.decode(t.off, dst, 2, func(b []byte, i int) float32 {
			return halfToF32(binary.LittleEndian.Uint16(b[i*2:]))
		})
		if err != nil {
			return nil, err
		}
	case "F32":
		err := f.decode(t.off, dst, 4, func(b []byte, i int) float32 {
			return math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
		})
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("safetensors: tensor %q dtype %q unsupported for an F32 read (want F32/BF16/F16)", name, t.dtype)
	}
	return dst, nil
}

// decode reads n elements of width bytes from off in decodeChunk-sized chunks and
// widens each into dst. dst is written exactly once per element, and no buffer
// larger than decodeChunk*width exists at any time.
func (f *safetensorsFile) decode(off int64, dst []float32, width int, widen func(b []byte, i int) float32) error {
	buf := make([]byte, min(decodeChunk, len(dst))*width)
	for i := 0; i < len(dst); {
		n := min(decodeChunk, len(dst)-i)
		b := buf[:n*width]
		if _, err := readFullAt(f.r, b, f.base+off+int64(i)*int64(width)); err != nil {
			return err
		}
		for j := range n {
			dst[i+j] = widen(b, j)
		}
		i += n
	}
	return nil
}

// Close releases the underlying file. Tensor bytes already returned stay valid:
// nothing aliases the file.
func (f *safetensorsFile) Close() error {
	f.closed = true
	return f.f.Close()
}

// halfToF32 expands an IEEE-754 half to float32, the same expansion aikit's
// embed.Float16sToF32 performs.
func halfToF32(h uint16) float32 {
	sign := uint32(h>>15) << 31
	exp := uint32(h>>10) & 0x1f
	mant := uint32(h) & 0x3ff
	switch exp {
	case 0:
		if mant == 0 {
			return math.Float32frombits(sign)
		}
		// Subnormal half: renormalise into a float32 exponent.
		e := uint32(127 - 15 + 1)
		for mant&0x400 == 0 {
			mant <<= 1
			e--
		}
		mant &= 0x3ff
		return math.Float32frombits(sign | e<<23 | mant<<13)
	case 0x1f:
		return math.Float32frombits(sign | 0xff<<23 | mant<<13)
	default:
		return math.Float32frombits(sign | (exp+127-15)<<23 | mant<<13)
	}
}

// readFullAt is io.ReadFull over io.ReaderAt: ReadAt may return a short read with
// its own error, and a short read with nil error is still a failure.
func readFullAt(r io.ReaderAt, b []byte, off int64) (int, error) {
	n, err := r.ReadAt(b, off)
	if n == len(b) {
		return n, nil
	}
	if err == nil {
		err = io.ErrUnexpectedEOF
	}
	return n, err
}

func shapeEqual(got, want []int) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
