package web

import (
	"errors"
	"net/http"

	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

// mapDomainError converts a domain error into the status code and body code it
// deserves. It is the single source of truth for that mapping: every caller
// that turns an error into a response must go through it, otherwise a domain
// error reported somewhere else loses its specific status.
//
// An error nobody anticipated falls through to 500 INTERNAL_ERROR, which
// names an unmapped failure honestly instead of borrowing a specific code.
func mapDomainError(err error) (int, server.ErrorCode) {
	var secErr *ogenerrors.SecurityError

	switch {
	// A request that never reached a handler, either because the header is absent
	// or because the handler refused the token. ogen reports the first case as a
	// bare ErrSecurityRequirementIsNotSatisfied, which wraps nothing, so it needs
	// its own check; the second arrives as a SecurityError wrapping the handler's
	// error and is spelled out separately for the same reason.
	case errors.As(err, &secErr):
		return http.StatusUnauthorized, server.ErrorCodeUNAUTHORIZED
	case errors.Is(err, ogenerrors.ErrSecurityRequirementIsNotSatisfied):
		return http.StatusUnauthorized, server.ErrorCodeUNAUTHORIZED
	case errors.Is(err, apperrors.ErrDBConvertError):
		return http.StatusInternalServerError, server.ErrorCodeDBCONVERTERROR
	case errors.Is(err, apperrors.ErrFactDoesNotExist):
		return http.StatusNotFound, server.ErrorCodeFACTDOESNOTEXIST
	case errors.Is(err, apperrors.ErrAlreadyHasFact):
		return http.StatusConflict, server.ErrorCodeALREADYHASFACT
	case errors.Is(err, apperrors.ErrEmptyEmbeddings),
		errors.Is(err, apperrors.ErrEmbeddingsQuantityNotEqualToFactsQuantity):
		return http.StatusUnprocessableEntity, server.ErrorCodeEMPTYEMBEDDINGS
	case errors.Is(err, apperrors.ErrGetEmbeddings),
		errors.Is(err, apperrors.ErrSummarize):
		return http.StatusBadGateway, server.ErrorCodeCANNOTGETMODELS
	case errors.Is(err, apperrors.ErrUserNotFound):
		return http.StatusUnauthorized, server.ErrorCodeUNAUTHORIZED
	case errors.Is(err, apperrors.ErrNoUsersConfigured):
		// Returned only while starting up, so it should never reach HTTP. If it
		// ever does, the server refused to start for want of a configured user:
		// a misconfiguration on our side, never the caller's fault.
		return http.StatusInternalServerError, server.ErrorCodeINTERNALERROR
	}

	return http.StatusInternalServerError, server.ErrorCodeINTERNALERROR
}
