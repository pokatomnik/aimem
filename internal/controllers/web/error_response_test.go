package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ht "github.com/ogen-go/ogen/http"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/validate"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
)

func TestRequestErrorHandlerRequestError(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "decode request",
			err: &ogenerrors.DecodeRequestError{
				OperationContext: ogenerrors.OperationContext{Name: "AddFact", ID: "addFact"},
				Err:              errors.New("unexpected token"),
			},
		},
		{
			name: "wrapped decode request",
			err: fmt.Errorf("handle: %w", &ogenerrors.DecodeRequestError{
				OperationContext: ogenerrors.OperationContext{Name: "AddFact", ID: "addFact"},
				Err:              errors.New("unexpected token"),
			}),
		},
		{
			name: "decode params",
			err: &ogenerrors.DecodeParamsError{
				OperationContext: ogenerrors.OperationContext{Name: "GetFactByID", ID: "getFactByID"},
				Err:              errors.New("bad param"),
			},
		},
		{
			name: "invalid content type",
			err:  validate.InvalidContentType("text/plain"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			RequestErrorHandler(context.Background(), rec, httptest.NewRequest(http.MethodPost, "/api/v1/fact", nil), tt.err)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}

			body := decodeAppError(t, rec)
			if body.Error != server.ErrorCodeBADREQUEST {
				t.Errorf("error = %q, want %q", body.Error, server.ErrorCodeBADREQUEST)
			}
			if body.Message == nil || *body.Message != tt.err.Error() {
				t.Errorf("message = %v, want %q", body.Message, tt.err.Error())
			}
		})
	}
}

// TestRequestErrorHandlerServerFault checks that a failure on our side is not
// reported as 400: doing so would hide a bug behind a client error.
func TestRequestErrorHandlerServerFault(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "encode failure",
			err:  errors.New("write: broken pipe"),
		},
		{
			name: "not implemented",
			err:  ht.ErrNotImplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			RequestErrorHandler(context.Background(), rec, httptest.NewRequest(http.MethodGet, "/api/v1/fact/x", nil), tt.err)

			if rec.Code != http.StatusInternalServerError {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
			}

			body := decodeAppError(t, rec)
			if body.Error != server.ErrorCodeINTERNALERROR {
				t.Errorf("error = %q, want %q", body.Error, server.ErrorCodeINTERNALERROR)
			}
		})
	}
}

func TestNotFoundHandler(t *testing.T) {
	rec := httptest.NewRecorder()

	NotFoundHandler(rec, httptest.NewRequest(http.MethodGet, "/api/v1/nope", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	body := decodeAppError(t, rec)
	if body.Error != server.ErrorCodeNOTFOUND {
		t.Errorf("error = %q, want %q", body.Error, server.ErrorCodeNOTFOUND)
	}
	if body.Message == nil || !strings.Contains(*body.Message, "/api/v1/nope") {
		t.Errorf("message = %v, want it to mention the requested path", body.Message)
	}
}

func TestWriteAppErrorContentType(t *testing.T) {
	rec := httptest.NewRecorder()

	writeAppError(rec, http.StatusConflict, server.ErrorCodeALREADYHASFACT, "duplicate")

	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json; charset=utf-8")
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}

	body := decodeAppError(t, rec)
	if body.Error != server.ErrorCodeALREADYHASFACT {
		t.Errorf("error = %q, want %q", body.Error, server.ErrorCodeALREADYHASFACT)
	}
	if body.Message == nil || *body.Message != "duplicate" {
		t.Errorf("message = %v, want %q", body.Message, "duplicate")
	}
}

// TestWriteAppErrorOmitsMessageOnlyWhenEmpty documents that an empty message
// still serialises: NewOptString sets Set even for "". Callers should pass
// something meaningful rather than rely on the field being dropped.
func TestWriteAppErrorEmptyMessage(t *testing.T) {
	rec := httptest.NewRecorder()

	writeAppError(rec, http.StatusBadRequest, server.ErrorCodeBADREQUEST, "")

	body := decodeAppError(t, rec)
	if body.Message == nil {
		t.Fatal("message is absent, want an empty string")
	}
	if *body.Message != "" {
		t.Errorf("message = %q, want empty", *body.Message)
	}
}

// TestServerOptionsApply wires the options into the generated server the way
// controller.go does, so a nil or rejected option fails here rather than at
// runtime.
func TestServerOptionsApply(t *testing.T) {
	_, err := server.NewServer(&Methods{}, ServerOptions()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
}

// decodeAppError unmarshals the recorded body into the appError shape.
func decodeAppError(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Error   server.ErrorCode `json:"error"`
	Message *string          `json:"message"`
} {
	t.Helper()

	var body struct {
		Error   server.ErrorCode `json:"error"`
		Message *string          `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}

	return body
}
