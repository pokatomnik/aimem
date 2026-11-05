package web

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

// mapDomainError is the single place an error becomes a status and a body code.
// The generated server routes handler errors to cfg.ErrorHandler rather than to a
// per-operation method, so this function — reached through RequestErrorHandler —
// is the only mapping that runs.
func TestMapDomainError(t *testing.T) {
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
		{
			name:       "user not found is unauthorized",
			err:        apperrors.ErrUserNotFound,
			wantStatus: http.StatusUnauthorized,
			wantCode:   server.ErrorCodeUNAUTHORIZED,
		},
		{
			name:       "no users configured is a server fault",
			err:        apperrors.ErrNoUsersConfigured,
			wantStatus: http.StatusInternalServerError,
			wantCode:   server.ErrorCodeINTERNALERROR,
		},
		{
			name: "a rejected token keeps its status through the wrapper",
			err: &ogenerrors.SecurityError{
				OperationContext: ogenerrors.OperationContext{Name: "AddFact", ID: "addFact"},
				Security:         "BearerAuth",
				Err:              apperrors.ErrUserNotFound,
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   server.ErrorCodeUNAUTHORIZED,
		},
		{
			name: "a missing header is unauthorized too",
			err: &ogenerrors.SecurityError{
				OperationContext: ogenerrors.OperationContext{Name: "AddFact", ID: "addFact"},
				Err:              ogenerrors.ErrSecurityRequirementIsNotSatisfied,
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   server.ErrorCodeUNAUTHORIZED,
		},
		{
			name: "an unanticipated error is an honest 500",
			err:  errors.New("something nobody anticipated"),
			// Not DB_CONVERT_ERROR or any other specific code: that would claim a
			// cause nobody established.
			wantStatus: http.StatusInternalServerError,
			wantCode:   server.ErrorCodeINTERNALERROR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code := mapDomainError(tt.err)

			if status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}

			if code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
		})
	}
}

// Repositories wrap domain errors with fmt.Errorf("%w"), so a tagged `switch err`
// or `err == apperrors.ErrX` would silently fall through to the default branch.
func TestMapDomainErrorMatchesWrappedErrors(t *testing.T) {
	inner := fmt.Errorf("get fact: %w", apperrors.ErrFactDoesNotExist)
	wrapped := fmt.Errorf("load: %w", inner)

	status, code := mapDomainError(wrapped)

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", status, http.StatusNotFound)
	}

	if code != server.ErrorCodeFACTDOESNOTEXIST {
		t.Errorf("code = %q, want %q", code, server.ErrorCodeFACTDOESNOTEXIST)
	}
}

// Every error declared in apperrors must have a row here stating the response it
// deserves, so adding one to the package without deciding how it surfaces fails
// loudly: a domain error with no case falls through to the default and reports an
// unmapped failure.
func TestEveryDomainErrorIsMapped(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   server.ErrorCode
	}{
		{"ErrDBConvertError", apperrors.ErrDBConvertError, http.StatusInternalServerError, server.ErrorCodeDBCONVERTERROR},
		{"ErrAlreadyHasFact", apperrors.ErrAlreadyHasFact, http.StatusConflict, server.ErrorCodeALREADYHASFACT},
		{"ErrGetEmbeddings", apperrors.ErrGetEmbeddings, http.StatusBadGateway, server.ErrorCodeCANNOTGETMODELS},
		{"ErrSummarize", apperrors.ErrSummarize, http.StatusBadGateway, server.ErrorCodeCANNOTGETMODELS},
		{"ErrEmbeddingsQuantityNotEqualToFactsQuantity", apperrors.ErrEmbeddingsQuantityNotEqualToFactsQuantity, http.StatusUnprocessableEntity, server.ErrorCodeEMPTYEMBEDDINGS},
		{"ErrEmptyEmbeddings", apperrors.ErrEmptyEmbeddings, http.StatusUnprocessableEntity, server.ErrorCodeEMPTYEMBEDDINGS},
		{"ErrFactDoesNotExist", apperrors.ErrFactDoesNotExist, http.StatusNotFound, server.ErrorCodeFACTDOESNOTEXIST},
		{"ErrUserNotFound", apperrors.ErrUserNotFound, http.StatusUnauthorized, server.ErrorCodeUNAUTHORIZED},
		// A startup failure, so it never surfaces over HTTP; if it does, it is a
		// server misconfiguration and the honest generic 500 is still right.
		{"ErrNoUsersConfigured", apperrors.ErrNoUsersConfigured, http.StatusInternalServerError, server.ErrorCodeINTERNALERROR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code := mapDomainError(tt.err)

			if status != tt.wantStatus || code != tt.wantCode {
				t.Errorf("mapDomainError(%s) = %d %q, want %d %q", tt.name, status, code, tt.wantStatus, tt.wantCode)
			}
		})
	}
}
