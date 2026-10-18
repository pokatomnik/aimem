package apperrors

import "errors"

var (
	// DB errors
	ErrorDBRead         = errors.New("db read error")
	ErrorDBWrite        = errors.New("db write error")
	ErrorDBConvertError = errors.New("db convert error")

	// Ollama client errors
	ErrorCannotGetModels = errors.New("ollama cannot get models")
	ErrorGetEmbeddings   = errors.New("ollama cannot get embeddings")

	// Business logic errors
	ErrorEmbeddingsQuantityNotEqualToFactsQuantity = errors.New("embeddings quantity does not equal to facts quantity")
	ErrorEmptyEmbeddings                           = errors.New("empty embeddings list")
)
