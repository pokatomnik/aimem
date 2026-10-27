package qdrant

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestHashMatchesKVDigest(t *testing.T) {
	t.Parallel()

	// The same construction as kv.hash: the hex-encoded SHA-256 of the text.
	const content = "The production database is hosted in Frankfurt"

	digest := sha256.Sum256([]byte(content))
	want := hex.EncodeToString(digest[:])

	got := hash(content)

	if got != want {
		t.Fatalf("hash() = %q, want %q", got, want)
	}

	if len(got) != sha256HexLength {
		t.Fatalf("hash length = %d, want %d", len(got), sha256HexLength)
	}
}

func TestFormatUUIDShapesDigest(t *testing.T) {
	t.Parallel()

	const digest = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	got := formatUUID(digest)
	want := "01234567-89ab-cdef-0123-456789abcdef"

	if got != want {
		t.Fatalf("formatUUID() = %q, want %q", got, want)
	}

	if len(got) != uuidDashed {
		t.Fatalf("UUID length = %d, want %d", len(got), uuidDashed)
	}
}

func TestFormatUUIDKeepsPrefixOnly(t *testing.T) {
	t.Parallel()

	// Two facts whose digests share the first 32 hex characters map to the same
	// point. This documents the collision bound of the design rather than hiding
	// it: 128 bits from a SHA-256 prefix.
	first := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa0000000000000000000000000000000f"
	second := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa11111111111111111111111111111111"

	if formatUUID(first) != formatUUID(second) {
		t.Fatalf("expected identical UUIDs for a shared 32-character prefix")
	}
}

func TestPointIDFromExternal(t *testing.T) {
	t.Parallel()

	const digest = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	const dashed = "01234567-89ab-cdef-0123-456789abcdef"

	tests := []struct {
		name string
		id   string
		want string
	}{
		{name: "canonical UUID", id: dashed, want: dashed},
		{name: "UUID without hyphens", id: digest[:uuidLength], want: dashed},
		{name: "full SHA-256 digest", id: digest, want: dashed},
		{name: "uppercase digest", id: "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF", want: dashed},
		{name: "surrounding spaces", id: "  " + dashed + "  ", want: dashed},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := pointIDFromExternal(test.id)
			if err != nil {
				t.Fatalf("pointIDFromExternal(%q) returned error: %v", test.id, err)
			}

			if got != test.want {
				t.Fatalf("pointIDFromExternal(%q) = %q, want %q", test.id, got, test.want)
			}
		})
	}
}

func TestPointIDFromExternalRejectsUnknownIDs(t *testing.T) {
	t.Parallel()

	ids := []string{
		"",
		"not-an-id",
		"0123456789abcdef0123456789abcde",     // 31 characters: one short of a UUID
		"0123456789abcdef0123456789abcdef0",   // 33 characters
		"01234567-89ab-cdef-0123-456789abcde", // dashed but 35 characters
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcde", // 63 characters
		"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",                                // right length, not hex
	}

	for _, id := range ids {
		t.Run(id, func(t *testing.T) {
			t.Parallel()

			if _, err := pointIDFromExternal(id); err == nil {
				t.Fatalf("pointIDFromExternal(%q) returned no error", id)
			}
		})
	}
}

func TestNewPointIDIsDerivedFromContent(t *testing.T) {
	t.Parallel()

	first := newPointID("same text")
	second := newPointID("same text")
	other := newPointID("other text")

	if first != second {
		t.Fatalf("newPointID is not deterministic: %q vs %q", first, second)
	}

	if first == other {
		t.Fatalf("different texts produced the same point ID")
	}

	if len(first) != uuidDashed {
		t.Fatalf("point ID %q has length %d, want %d", first, len(first), uuidDashed)
	}

	if _, err := pointIDFromExternal(first); err != nil {
		t.Fatalf("point ID %q is not accepted back: %v", first, err)
	}
}
