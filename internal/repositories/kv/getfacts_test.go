package kv

// TODO: GetFactsList is untested — neither the distance-to-similarity
// conversion (xvec reports cosine as a distance, 0 = identical) nor the
// threshold and topK handling is covered.
//
// The seam now exists: `collection` in kv.go is satisfied by xvec.Collection and
// by a fake, and getfact_test.go shows the pattern (newFakeKV). Cover the logic
// of this package through it instead of opening a real xvec collection, which
// would exercise the library rather than our code and break the hermetic-test
// rules (no filesystem, no I/O).
//
// Assert on our code only — the "similarity := 1 - result.Score" inversion,
// filtering by similarityThreshold, topK as an upper bound applied before
// filtering, and ErrDBConvertError when a projected field is not a string.
