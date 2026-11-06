package apperrors

import (
	"errors"
	"fmt"
	"testing"
)

// The log severity rule and the HTTP status table are separate, but both have to
// agree on which finished sentinels are the caller's fault. Pinning them here
// means a new error cannot quietly become a server fault in the log.
func TestIsClientError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"user not found", ErrUserNotFound, true},
		{"fact does not exist", ErrFactDoesNotExist, true},
		{"already has fact", ErrAlreadyHasFact, true},
		{"empty embeddings", ErrEmptyEmbeddings, true},
		{"embeddings quantity mismatch", ErrEmbeddingsQuantityNotEqualToFactsQuantity, true},
		{"db convert error", ErrDBConvertError, false},
		{"cannot get embeddings", ErrGetEmbeddings, false},
		{"cannot summarize", ErrSummarize, false},
		{"no users configured", ErrNoUsersConfigured, false},
		{"an unknown error", errors.New("boom"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsClientError(tt.err); got != tt.want {
				t.Errorf("IsClientError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// Repositories wrap sentinels with %w, so the check has to follow the chain;
// comparing by identity would classify every wrapped refusal as a server fault.
func TestIsClientErrorFollowsTheChain(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("load: %w", fmt.Errorf("get fact: %w", ErrFactDoesNotExist))

	if !IsClientError(wrapped) {
		t.Errorf("IsClientError(%v) = false, want true", wrapped)
	}
}
