package qdrant

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// fakeConfig satisfies the package's config interface without reading any file or
// environment variable.
type fakeConfig struct {
	host       string
	apiKey     string
	dimensions uint32
}

func (f fakeConfig) GetQdrantHost() string   { return f.host }
func (f fakeConfig) GetQdrantAPIKey() string { return f.apiKey }
func (f fakeConfig) GetDimensions() uint32   { return f.dimensions }

// recordedRequest is one request the repository made.
type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Header http.Header
	Body   []byte
}

// fakeTransport answers requests from a handler and records what it saw. It
// keeps the tests off the network and off the filesystem.
type fakeTransport struct {
	requests []recordedRequest
	handle   func(request recordedRequest) (status int, body string)
}

func (f *fakeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	var body []byte

	// A request without a body has a nil Body, not an empty one.
	if request.Body != nil {
		var err error

		body, err = io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
	}

	recorded := recordedRequest{
		Method: request.Method,
		Path:   request.URL.Path,
		Query:  request.URL.RawQuery,
		Header: request.Header.Clone(),
		Body:   body,
	}
	f.requests = append(f.requests, recorded)

	status, responseBody := http.StatusOK, "{}"
	if f.handle != nil {
		status, responseBody = f.handle(recorded)
	}

	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
		Header:     make(http.Header),
	}, nil
}

// newTestRepository builds a repository backed by a fake transport and records
// every request for later assertions.
func newTestRepository(handle func(request recordedRequest) (int, string)) (*Qdrant, *fakeTransport) {
	transport := &fakeTransport{handle: handle}

	return New(fakeConfig{
		host:       "http://qdrant.test",
		dimensions: 3,
	}, &http.Client{Transport: transport}), transport
}

// collectionHandlers answers the collection bootstrap that every operation runs
// first, delegating the request that matters to the given handler.
func collectionHandlers(handle func(request recordedRequest) (int, string)) func(recordedRequest) (int, string) {
	return func(request recordedRequest) (int, string) {
		if request.Path == pathCollectionExists {
			return http.StatusOK, `{"result":{"exists":true},"status":"ok"}`
		}

		return handle(request)
	}
}

func mustDecodeBody(t interface{ Fatalf(string, ...any) }, body []byte, out any) {
	if err := json.Unmarshal(body, out); err != nil {
		t.Fatalf("decode request body %q: %v", body, err)
	}
}
