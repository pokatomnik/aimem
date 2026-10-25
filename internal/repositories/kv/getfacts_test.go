package kv

// TODO: GetFactsList is untested — neither the distance-to-similarity
// conversion (xvec reports cosine as a distance, 0 = identical) nor the
// threshold and topK handling is covered.
//
// Do not cover it by driving a real xvec collection in t.TempDir(), the way
// the previous test did: that exercises the library, which is already tested
// upstream, and it breaks the hermetic-test rules (no filesystem, no I/O).
//
// Cover the logic of this package instead: extract the collection behind a
// narrow unexported interface next to its consumer, fake it, and feed it
// scores and fields directly. Then assert on our code only — the
// "similarity := 1 - result.Score" inversion, filtering by
// similarityThreshold, topK as an upper bound applied before filtering, and
// ErrDBConvertError when a projected field is not a string.
