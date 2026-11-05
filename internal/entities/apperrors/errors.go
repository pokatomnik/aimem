package apperrors

import "errors"

var (
	// DB errors
	ErrDBConvertError = errors.New("db convert error")
	ErrAlreadyHasFact = errors.New("db already has fact")

	// LLM client errors
	ErrGetEmbeddings = errors.New("cannot get embeddings")
	ErrSummarize     = errors.New("cannot summarize")

	// Business logic errors
	ErrEmbeddingsQuantityNotEqualToFactsQuantity = errors.New("embeddings quantity does not equal to facts quantity")
	ErrEmptyEmbeddings                           = errors.New("empty embeddings list")
	ErrFactDoesNotExist                          = errors.New("document does not exist")

	// Users errors
	ErrUserNotFound      = errors.New("unauthorized")
	ErrNoUsersConfigured = errors.New("no users configured")
)
