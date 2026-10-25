package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

func TestNewErrorMapsDomainErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   server.ErrorCode
	}{
		{
			name:       "db convert error",
			err:        apperrors.ErrDBConvertError,
			wantStatus: http.StatusInternalServerError,
			wantCode:   server.ErrorCodeDBCONVERTERROR,
		},
		{
			name:       "fact does not exist",
			err:        apperrors.ErrFactDoesNotExist,
			wantStatus: http.StatusNotFound,
			wantCode:   server.ErrorCodeFACTDOESNOTEXIST,
		},
		{
			name:       "already has fact",
			err:        apperrors.ErrAlreadyHasFact,
			wantStatus: http.StatusConflict,
			wantCode:   server.ErrorCodeALREADYHASFACT,
		},
		{
			name:       "empty embeddings",
			err:        apperrors.ErrEmptyEmbeddings,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   server.ErrorCodeEMPTYEMBEDDINGS,
		},
		{
			name:       "embeddings quantity mismatch",
			err:        apperrors.ErrEmbeddingsQuantityNotEqualToFactsQuantity,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   server.ErrorCodeEMPTYEMBEDDINGS,
		},
		{
			name:       "cannot get models",
			err:        apperrors.ErrCannotGetModels,
			wantStatus: http.StatusBadGateway,
			wantCode:   server.ErrorCodeCANNOTGETMODELS,
		},
		{
			name:       "get embeddings",
			err:        apperrors.ErrGetEmbeddings,
			wantStatus: http.StatusBadGateway,
			wantCode:   server.ErrorCodeCANNOTGETMODELS,
		},
		{
			name:       "summarize",
			err:        apperrors.ErrSummarize,
			wantStatus: http.StatusBadGateway,
			wantCode:   server.ErrorCodeCANNOTGETMODELS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := (&Methods{}).NewError(context.Background(), tt.err)

			if got == nil {
				t.Fatal("NewError returned nil: encodeErrorResponse dereferences StatusCode and would panic")
			}
			if got.StatusCode != tt.wantStatus {
				t.Errorf("StatusCode = %d, want %d", got.StatusCode, tt.wantStatus)
			}
			if got.Response.Error != tt.wantCode {
				t.Errorf("Error = %q, want %q", got.Response.Error, tt.wantCode)
			}
			if !got.Response.Message.Set {
				t.Error("Message is unset, want the error text")
			}
			if got.Response.Message.Value != tt.err.Error() {
				t.Errorf("Message = %q, want %q", got.Response.Message.Value, tt.err.Error())
			}
		})
	}
}

// TestNewErrorMatchesWrappedErrors guards the errors.Is form of the switch.
// Repositories wrap domain errors with fmt.Errorf("%w"), so a tagged
// `switch err` or `err == apperrors.ErrX` would silently fall through to the
// default branch.
func TestNewErrorMatchesWrappedErrors(t *testing.T) {
	inner := fmt.Errorf("get fact: %w", apperrors.ErrFactDoesNotExist)
	wrapped := fmt.Errorf("load: %w", inner)

	got := (&Methods{}).NewError(context.Background(), wrapped)

	if got.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", got.StatusCode, http.StatusNotFound)
	}
	if got.Response.Error != server.ErrorCodeFACTDOESNOTEXIST {
		t.Errorf("Error = %q, want %q", got.Response.Error, server.ErrorCodeFACTDOESNOTEXIST)
	}
	if got.Response.Message.Value != wrapped.Error() {
		t.Errorf("Message = %q, want the full wrapped text %q", got.Response.Message.Value, wrapped.Error())
	}
}

// TestNewErrorUnknownError pins the default branch: an error nobody
// anticipated is reported as 500 with INTERNAL_ERROR, which names an unmapped
// failure honestly. It must not borrow DB_CONVERT_ERROR — that code names a
// specific cause and now has its own case.
func TestNewErrorUnknownError(t *testing.T) {
	err := errors.New("something nobody anticipated")

	got := (&Methods{}).NewError(context.Background(), err)

	if got.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", got.StatusCode, http.StatusInternalServerError)
	}
	if got.Response.Error != server.ErrorCodeINTERNALERROR {
		t.Errorf("Error = %q, want %q", got.Response.Error, server.ErrorCodeINTERNALERROR)
	}
}
