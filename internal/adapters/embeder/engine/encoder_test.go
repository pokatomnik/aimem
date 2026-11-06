package engine

import (
	"testing"
	"testing/fstest"
)

// tinyTokenizer is a self-contained tokenizer.json in the shape
// tokenizer.LoadJSONBytes accepts: a BPE model with an explicit vocab and no
// byte-level pipeline, so encoding is a plain longest-match over the pieces. It
// is only here to exercise the wiring — which ids a string becomes is the
// tokenizer package's business, and the production checkpoint's own tokenizer is
// checked from the package that embeds it.
//
// <bos> and <eos> carry the ids the tiny fixture's model uses (2 and 1), so the
// wrapped stream stays inside that model's 256-entry vocabulary.
const tinyTokenizer = `{
  "added_tokens": [
    {"id": 2, "content": "<bos>", "rstrip": false},
    {"id": 1, "content": "<eos>", "rstrip": false}
  ],
  "model": {
    "type": "BPE",
    "vocab": {"<unk>": 0, "<eos>": 1, "<bos>": 2, "a": 3, "b": 4, "ab": 5, "\u2581": 6, "\u2581a": 7, "\u2581ab": 8},
    "merges": [["a", "b"]]
  },
  "normalizer": null,
  "pre_tokenizer": null,
  "decoder": null,
  "post_processor": null
}`

// encoderFixture is the tiny checkpoint plus tinyTokenizer, so LoadEncoder has all
// three files it reads.
func encoderFixture(t *testing.T) fstest.MapFS {
	t.Helper()
	mfs := fstest.MapFS{}
	for _, name := range []string{"config.json", "model.safetensors"} {
		raw, err := goldenFS.ReadFile(goldenDir + "/" + name)
		if err != nil {
			t.Fatalf("read fixture %s: %v", name, err)
		}
		mfs[name] = &fstest.MapFile{Data: raw}
	}
	mfs["tokenizer.json"] = &fstest.MapFile{Data: []byte(tinyTokenizer)}
	return mfs
}

// TestLoadEncoderFromFS covers the wiring this package owns: model, tokenizer,
// bos/eos lookup and the EmbedText pipeline, all from an fs.FS with nothing on
// disk.
func TestLoadEncoderFromFS(t *testing.T) {
	enc, err := LoadEncoder(encoderFixture(t), ".")
	if err != nil {
		t.Fatalf("LoadEncoder: %v", err)
	}
	if enc.m == nil || enc.tok == nil {
		t.Fatal("encoder is not fully wired")
	}
	if enc.bos != 2 || enc.eos != 1 {
		t.Fatalf("bos/eos = %d/%d, want 2/1 from the checkpoint's tokenizer", enc.bos, enc.eos)
	}

	ids, err := enc.tokenize("ab")
	if err != nil {
		t.Fatalf("tokenize: %v", err)
	}
	if len(ids) < 3 {
		t.Fatalf("ids %v: want at least bos+one piece+eos", ids)
	}
	if ids[0] != enc.bos || ids[len(ids)-1] != enc.eos {
		t.Fatalf("ids %v: want %d first and %d last", ids, enc.bos, enc.eos)
	}
	for _, id := range ids {
		if id < 0 || id >= 256 {
			t.Fatalf("id %d outside the tiny model's 256-entry vocabulary", id)
		}
	}
}

// TestEmbedTextIsTokenizeThenEmbed pins the pipeline: EmbedText must be exactly
// the model embedding of the wrapped ids, so a change in either half cannot pass
// unnoticed.
func TestEmbedTextIsTokenizeThenEmbed(t *testing.T) {
	enc, err := LoadEncoder(encoderFixture(t), ".")
	if err != nil {
		t.Fatalf("LoadEncoder: %v", err)
	}

	ids, err := enc.tokenize("ab")
	if err != nil {
		t.Fatalf("tokenize: %v", err)
	}
	want, err := enc.m.embedIDs(ids)
	if err != nil {
		t.Fatalf("embedIDs: %v", err)
	}
	got, err := enc.EmbedText("ab")
	if err != nil {
		t.Fatalf("EmbedText: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("EmbedText returned %d values, embedIDs %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("value %d: EmbedText %v != embedIDs %v", i, got[i], want[i])
		}
	}
}

// TestLoadEncoderNeedsTokenizer covers the one file LoadEncoder reads itself:
// a checkpoint without tokenizer.json must say so instead of failing later.
func TestLoadEncoderNeedsTokenizer(t *testing.T) {
	mfs := encoderFixture(t)
	delete(mfs, "tokenizer.json")
	if _, err := LoadEncoder(mfs, "."); err == nil {
		t.Fatal("LoadEncoder succeeded without tokenizer.json")
	}
}
