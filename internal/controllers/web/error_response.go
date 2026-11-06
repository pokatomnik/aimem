package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/validate"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
)

// ServerOptions makes the generated server answer with appError JSON instead
// of ogen's text/plain defaults. Pass them to server.NewServer with the process
// logger:
//
//	server.NewServer(methods, security, web.ServerOptions(logger)...)
func ServerOptions(l logger) []server.ServerOption {
	return []server.ServerOption{
		server.WithErrorHandler(requestErrorHandler(l)),
		server.WithNotFound(notFoundHandler(l)),
	}
}

// requestErrorHandler logs a failure and then answers exactly as
// RequestErrorHandler does.
//
// Logging lives here, not in the operation handlers, so every error that becomes
// a response is recorded once, with the method and path that only the HTTP edge
// knows. A failure on our side is an error; a request we refuse is a warning.
func requestErrorHandler(l logger) server.ErrorHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		status, code := classifyError(err)
		logRequestError(l, r, status, code, err)

		RequestErrorHandler(ctx, w, r, err)
	}
}

// notFoundHandler logs a request the spec does not describe. It is a warning:
// the caller asked for a route this server never had.
func notFoundHandler(l logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if l != nil {
			l.Warn("route not found", "method", r.Method, "path", r.URL.Path, "status", http.StatusNotFound)
		}

		NotFoundHandler(w, r)
	}
}

// classifyError picks the status and body code for err. It is the single
// classification used both to answer and to log, so the two cannot disagree
// about whether a failure was the caller's fault.
func classifyError(err error) (int, server.ErrorCode) {
	if isRequestError(err) {
		return http.StatusBadRequest, server.ErrorCodeBADREQUEST
	}

	return mapDomainError(err)
}

// logRequestError records a failed request at the severity its status deserves.
// A nil logger is tolerated so the handlers can be exercised without one.
func logRequestError(l logger, r *http.Request, status int, code server.ErrorCode, err error) {
	if l == nil {
		return
	}

	args := []any{
		"method", r.Method,
		"path", r.URL.Path,
		"status", status,
		"code", code,
		"error", err,
	}

	if status >= http.StatusInternalServerError {
		l.Error("request failed", args...)

		return
	}

	l.Warn("request rejected", args...)
}

// RequestErrorHandler handles every failure the generated server cannot turn
// into a response, which includes three distinct kinds.
//
// Failures raised before a handler of ours runs — malformed JSON, a body or
// parameter that fails spec validation, an unexpected Content-Type — do not
// match the contract, so they are reported as 400.
//
// Authentication failures arrive here too: ogen reports a missing header as a
// bare ogenerrors.ErrSecurityRequirementIsNotSatisfied and a rejected token as a
// SecurityError wrapping our own error. mapDomainError turns both into 401.
//
// Errors returned by a handler arrive here too, because the generated server
// routes them to the configured error handler rather than to a per-operation
// NewError method. Those carry domain meaning, so they go through mapDomainError
// and keep their own status: a missing fact stays 404, a duplicate stays 409.
// Reporting them as a blanket 500 would erase that distinction.
func RequestErrorHandler(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	// A request error is 400; domain and unknown errors alike go through
	// mapDomainError, which picks the right status for a known cause and falls
	// back to 500 for anything nobody anticipated, so a server bug is never
	// reported as a client error.
	status, code := classifyError(err)
	writeAppError(w, status, code, err.Error())
}

// NotFoundHandler answers requests whose path is absent from the spec.
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	writeAppError(w, http.StatusNotFound, server.ErrorCodeNOTFOUND, "unknown route: "+r.URL.Path)
}

// isRequestError reports whether err is caused by the request itself rather
// than by our code.
func isRequestError(err error) bool {
	var (
		decodeRequest *ogenerrors.DecodeRequestError
		decodeParams  *ogenerrors.DecodeParamsError
		contentType   *validate.InvalidContentTypeError
	)

	return errors.As(err, &decodeRequest) ||
		errors.As(err, &decodeParams) ||
		errors.As(err, &contentType)
}

// writeAppError writes code with an appError body.
func writeAppError(w http.ResponseWriter, code int, ec server.ErrorCode, message string) {
	body, err := json.Marshal(server.AppError{
		Error:   ec,
		Message: server.NewOptString(message),
	})
	if err != nil {
		http.Error(w, http.StatusText(code), code)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write(body)
}
