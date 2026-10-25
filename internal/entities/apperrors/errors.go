package apperrors

import "errors"

var (
	// DB errors
	ErrDBConvertError = errors.New("db convert error")
	ErrAlreadyHasFact = errors.New("db already has fact")

	// Ollama client errors
	ErrCannotGetModels = errors.New("ollama cannot get models")
	ErrGetEmbeddings   = errors.New("ollama cannot get embeddings")
	ErrSummarize       = errors.New("ollama cannot summarize")

	// Business logic errors
	ErrEmbeddingsQuantityNotEqualToFactsQuantity = errors.New("embeddings quantity does not equal to facts quantity")
	ErrEmptyEmbeddings                           = errors.New("empty embeddings list")
	ErrFactDoesNotExist                          = errors.New("document does not exist")
)
