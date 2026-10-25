package web

import (
	"context"
	"errors"
	"net/http"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

// NewError maps domain errors to the common default response.
//
// It is called by the generated server for every error returned by a handler,
// so it is the single place where apperrors become HTTP status codes.
func (m *Methods) NewError(ctx context.Context, err error) *server.AppErrorStatusCode {
	var (
		code = http.StatusInternalServerError
		ec   = server.ErrorCodeINTERNALERROR
	)

	switch {
	case errors.Is(err, apperrors.ErrDBConvertError):
		code, ec = http.StatusInternalServerError, server.ErrorCodeDBCONVERTERROR
	case errors.Is(err, apperrors.ErrFactDoesNotExist):
		code, ec = http.StatusNotFound, server.ErrorCodeFACTDOESNOTEXIST
	case errors.Is(err, apperrors.ErrAlreadyHasFact):
		code, ec = http.StatusConflict, server.ErrorCodeALREADYHASFACT
	case errors.Is(err, apperrors.ErrEmptyEmbeddings),
		errors.Is(err, apperrors.ErrEmbeddingsQuantityNotEqualToFactsQuantity):
		code, ec = http.StatusUnprocessableEntity, server.ErrorCodeEMPTYEMBEDDINGS
	case errors.Is(err, apperrors.ErrCannotGetModels),
		errors.Is(err, apperrors.ErrGetEmbeddings),
		errors.Is(err, apperrors.ErrSummarize):
		code, ec = http.StatusBadGateway, server.ErrorCodeCANNOTGETMODELS
	}

	return &server.AppErrorStatusCode{
		StatusCode: code,
		Response: server.AppError{
			Error:   ec,
			Message: server.NewOptString(err.Error()),
		},
	}
}
