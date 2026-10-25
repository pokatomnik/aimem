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
// of ogen's text/plain defaults. Pass them to server.NewServer:
//
//	server.NewServer(methods, web.ServerOptions()...)
func ServerOptions() []server.ServerOption {
	return []server.ServerOption{
		server.WithErrorHandler(RequestErrorHandler),
		server.WithNotFound(NotFoundHandler),
	}
}

// RequestErrorHandler handles every failure the generated server cannot turn
// into a response, which includes two distinct kinds.
//
// Failures raised before a handler of ours runs — malformed JSON, a body or
// parameter that fails spec validation, an unexpected Content-Type — do not
// match the contract, so they are reported as 400.
//
// Errors returned by a handler arrive here too, because the generated server
// routes them to the configured error handler rather than calling NewError.
// Those carry domain meaning, so they go through mapDomainError and keep their
// own status: a missing fact stays 404, a duplicate stays 409. Reporting them
// as a blanket 500 would erase that distinction.
func RequestErrorHandler(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	if isRequestError(err) {
		writeAppError(w, http.StatusBadRequest, server.ErrorCodeBADREQUEST, err.Error())
		return
	}

	// Domain and unknown errors alike: mapDomainError picks the right status for
	// a known cause and falls back to 500 for anything nobody anticipated, so a
	// server bug is never reported as a client error.
	code, ec := mapDomainError(err)
	writeAppError(w, code, ec, err.Error())
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
