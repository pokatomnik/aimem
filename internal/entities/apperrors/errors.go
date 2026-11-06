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

// IsClientError reports whether err is an expected refusal caused by the caller
// or by the data it supplied, rather than a fault in this process.
//
// It is the severity rule for logging at the edges: a client error is a warning,
// anything else is an error. The HTTP status table stays in the web controller,
// which has its own mapping; this answers only "whose fault was it", so the MCP
// tools and the HTTP API cannot disagree about which failures deserve an error
// entry.
func IsClientError(err error) bool {
	return errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrFactDoesNotExist) ||
		errors.Is(err, ErrAlreadyHasFact) ||
		errors.Is(err, ErrEmptyEmbeddings) ||
		errors.Is(err, ErrEmbeddingsQuantityNotEqualToFactsQuantity)
}
