package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type (
	// httpError is a failed request. StatusCode is the HTTP status and Message is
	// the server's own description, taken from the response envelope when it has
	// one.
	httpError struct {
		StatusCode int
		Message    string
	}
)

func (e *httpError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("qdrant: request failed with status %d", e.StatusCode)
	}

	return fmt.Sprintf("qdrant: request failed with status %d: %s", e.StatusCode, e.Message)
}

// do performs one request and decodes the envelope Qdrant wraps around every
// answer. A nil body means no request body; a nil out means the result is
// ignored. The whole response is read before parsing so a connection error
// surfaces as itself rather than as a JSON error.
func (q *Qdrant) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	baseURL := q.baseURL()
	if baseURL == "" {
		err := fmt.Errorf("qdrant: host is not configured")
		q.logger.Error("qdrant request failed", "method", method, "path", path, "error", err)

		return err
	}

	target := baseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}

	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			q.logger.Error("encode qdrant request failed", "method", method, "path", path, "error", err)

			return err
		}

		requestBody = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, target, requestBody)
	if err != nil {
		q.logger.Error("build qdrant request failed", "method", method, "path", path, "error", err)

		return err
	}

	if body != nil {
		request.Header.Set(headerContentType, contentTypeJSON)
	}

	if key := q.config.GetQdrantAPIKey(); key != "" {
		request.Header.Set(headerAPIKey, key)
	}

	response, err := q.client.Do(request)
	if err != nil {
		q.logger.Error("perform qdrant request failed", "method", method, "path", path, "error", err)

		return err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		q.logger.Error("read qdrant response failed", "method", method, "path", path, "status", response.StatusCode, "error", err)

		return err
	}

	if response.StatusCode != http.StatusOK {
		httpErr := &httpError{
			StatusCode: response.StatusCode,
			Message:    errorMessage(data),
		}
		q.logger.Error("qdrant request failed", "method", method, "path", path, "status", httpErr.StatusCode, "error", httpErr)

		return httpErr
	}

	if out == nil {
		return nil
	}

	var decoded envelope[json.RawMessage]
	if err := json.Unmarshal(data, &decoded); err != nil {
		q.logger.Error("decode qdrant response envelope failed", "method", method, "path", path, "status", response.StatusCode, "error", err)

		return err
	}

	if decoded.Result == nil {
		err := fmt.Errorf("qdrant: response has no result")
		q.logger.Error("qdrant response has no result", "method", method, "path", path, "error", err)

		return err
	}

	if err := json.Unmarshal(*decoded.Result, out); err != nil {
		q.logger.Error("decode qdrant response failed", "method", method, "path", path, "error", err)

		return err
	}

	return nil
}

// baseURL returns the configured host without trailing slashes, so paths can be
// concatenated directly.
func (q *Qdrant) baseURL() string {
	host := q.config.GetQdrantHost()
	for strings.HasSuffix(host, "/") {
		host = strings.TrimSuffix(host, "/")
	}

	return host
}

// writeQuery makes a write visible to the next read. Without it Qdrant may only
// acknowledge the operation, and a fact saved by remember would be missing from
// an immediate recall.
func writeQuery() url.Values {
	return url.Values{queryParamWait: []string{"true"}}
}

// errorMessage extracts the server's description from a failed response. A
// failure answers with status as an object, {"error": "..."}; a successful one
// answers with the string "ok", which is why the raw form is decoded here
// instead of a typed Status field.
func errorMessage(data []byte) string {
	var decoded struct {
		Status json.RawMessage `json:"status"`
	}

	if err := json.Unmarshal(data, &decoded); err == nil && len(decoded.Status) > 0 {
		var status struct {
			Error string `json:"error"`
		}

		if err := json.Unmarshal(decoded.Status, &status); err == nil && status.Error != "" {
			return status.Error
		}
	}

	return strings.TrimSpace(string(data))
}

func formatNumericID(id uint64) string {
	return strconv.FormatUint(id, 10)
}
