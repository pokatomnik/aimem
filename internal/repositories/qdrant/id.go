package qdrant

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

const (
	// uuidLength is the number of hex characters a UUID encodes (128 bits),
	// without the four hyphens.
	uuidLength = 32
	uuidDashed = 36

	// sha256HexLength is the length of the hex-encoded SHA-256 digest, which is
	// what kv uses as a primary key.
	sha256HexLength = 64
)

// hash returns the SHA-256 digest of the fact text, hex encoded. It is the same
// digest kv uses, so a fact has one identity regardless of the backend.
func hash(data string) string {
	digest := sha256.Sum256([]byte(data))

	return hex.EncodeToString(digest[:])
}

// newPointID derives the Qdrant point ID from a fact's text: the SHA-256 digest,
// as in kv, with its first 32 hex characters shaped into a UUID.
//
// Qdrant accepts a point ID only as an unsigned integer or a UUID, so the digest
// cannot be used verbatim. Truncating to 128 bits keeps the ID deterministic and
// stable across runs, and the UUID parser Qdrant uses does not check the version
// or variant bits, so an arbitrary prefix is a valid UUID as far as it cares.
func newPointID(content string) string {
	return formatUUID(hash(content))
}

// formatUUID turns the leading 32 hex characters of a digest into a canonical
// lowercase UUID. The caller must pass at least uuidLength characters.
func formatUUID(digest string) string {
	digest = strings.ToLower(digest[:uuidLength])

	return digest[0:8] + "-" + digest[8:12] + "-" + digest[12:16] + "-" + digest[16:20] + "-" + digest[20:32]
}

// pointIDFromExternal accepts the IDs this package hands out and maps them to the
// form Qdrant stores.
//
// GetFactsList returns UUIDs, so that is the shape callers pass back. A full
// SHA-256 digest (the kv identifier, or one printed by an older dataset) is
// accepted too and reduced the same way newPointID does it, so data carried over
// from kv resolves to the same point.
func pointIDFromExternal(id string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(id))

	switch len(normalized) {
	case uuidDashed:
		if !isHexWithHyphens(normalized) {
			return "", fmt.Errorf("%w: %q is not a UUID", apperrors.ErrDBConvertError, id)
		}

		return normalized, nil

	case uuidLength:
		if !isHex(normalized) {
			return "", fmt.Errorf("%w: %q is not a UUID", apperrors.ErrDBConvertError, id)
		}

		return formatUUID(normalized), nil

	case sha256HexLength:
		if !isHex(normalized) {
			return "", fmt.Errorf("%w: %q is not a SHA-256 digest", apperrors.ErrDBConvertError, id)
		}

		return formatUUID(normalized), nil
	}

	return "", fmt.Errorf("%w: %q is neither a UUID nor a SHA-256 digest", apperrors.ErrDBConvertError, id)
}

func isHex(value string) bool {
	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		default:
			return false
		}
	}

	return true
}

// isHexWithHyphens reports whether value is a UUID in canonical 8-4-4-4-12 form.
func isHexWithHyphens(value string) bool {
	if len(value) != uuidDashed {
		return false
	}

	for i, r := range value {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if r != '-' {
				return false
			}

			continue
		}

		if !isHex(string(r)) {
			return false
		}
	}

	return true
}
