package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
)

// spaTestFS mirrors the shape of a real build: an index naming hashed assets.
func spaTestFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte(`<!doctype html><script type="module" src="./assets/index-CT2NxLkI.js"></script>`),
		},
		"assets/index-CT2NxLkI.js":  &fstest.MapFile{Data: []byte("console.log(1)")},
		"assets/index-SLAhs_H4.css": &fstest.MapFile{Data: []byte("body{margin:0}")},
		"favicon.svg":               &fstest.MapFile{Data: []byte("<svg/>")},
		"unknown.bin":               &fstest.MapFile{Data: []byte{0x00, 0x01}},
	}
}

// spaRequest serves target with authentication off, which is the common case in
// these tests: the flag only changes index.html.
func spaRequest(t *testing.T, fsys fs.FS, method, target string) *httptest.ResponseRecorder {
	t.Helper()

	return spaRequestAuth(t, fsys, method, target, false)
}

func spaRequestAuth(t *testing.T, fsys fs.FS, method, target string, authRequired bool) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	newFileHandler(fsys, authRequired).ServeHTTP(rec, httptest.NewRequest(method, target, nil))

	return rec
}

func TestSPAServesIndexAtRoot(t *testing.T) {
	rec := spaRequest(t, spaTestFS(), http.MethodGet, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Body.String(); !strings.Contains(got, "index-CT2NxLkI.js") {
		t.Errorf("body does not contain the index contents: %q", got)
	}

	if got, want := rec.Header().Get("Content-Type"), "text/html; charset=utf-8"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}

	if got, want := rec.Header().Get("Cache-Control"), "no-cache"; got != want {
		t.Errorf("Cache-Control = %q, want %q", got, want)
	}
}

func TestSPAServesHashedAssetsWithLongCache(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		wantType string
		wantBody string
	}{
		{
			name:     "js",
			target:   "/assets/index-CT2NxLkI.js",
			wantType: "text/javascript; charset=utf-8",
			wantBody: "console.log(1)",
		},
		{
			name:     "css",
			target:   "/assets/index-SLAhs_H4.css",
			wantType: "text/css; charset=utf-8",
			wantBody: "body{margin:0}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := spaRequest(t, spaTestFS(), http.MethodGet, tt.target)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}

			if got := rec.Header().Get("Content-Type"); got != tt.wantType {
				t.Errorf("Content-Type = %q, want %q", got, tt.wantType)
			}

			if got, want := rec.Header().Get("Cache-Control"), "public, max-age=31536000, immutable"; got != want {
				t.Errorf("Cache-Control = %q, want %q", got, want)
			}

			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

// Files outside assets/ keep their name across builds, so they must revalidate.
func TestSPANonAssetFilesAreNotCachedForever(t *testing.T) {
	rec := spaRequest(t, spaTestFS(), http.MethodGet, "/favicon.svg")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got, want := rec.Header().Get("Cache-Control"), "no-cache"; got != want {
		t.Errorf("Cache-Control = %q, want %q", got, want)
	}
}

func TestSPAUnknownExtensionFallsBackToOctetStream(t *testing.T) {
	rec := spaRequest(t, spaTestFS(), http.MethodGet, "/unknown.bin")

	if got, want := rec.Header().Get("Content-Type"), "application/octet-stream"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}
}

func TestSPAContentLengthMatchesBody(t *testing.T) {
	rec := spaRequest(t, spaTestFS(), http.MethodGet, "/assets/index-SLAhs_H4.css")

	if got, want := rec.Header().Get("Content-Length"), "14"; got != want {
		t.Errorf("Content-Length = %q, want %q", got, want)
	}

	if got := rec.Body.Len(); got != 14 {
		t.Errorf("body length = %d, want 14", got)
	}
}

func TestSPAMissingPathIsNotFound(t *testing.T) {
	// No SPA fallback: HashRouter keeps the route after '#', so the server only
	// sees "/" and a missing path really is missing.
	tests := []string{"/missing.html", "/assets/missing-abc.js", "/assets/"}

	for _, target := range tests {
		t.Run(target, func(t *testing.T) {
			rec := spaRequest(t, spaTestFS(), http.MethodGet, target)

			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}
		})
	}
}

func TestSPAPathTraversalIsRejected(t *testing.T) {
	tests := []string{
		"/../spa.go",
		"/assets/../../spa.go",
		"/%2e%2e/%2e%2e/spa.go",
	}

	for _, target := range tests {
		t.Run(target, func(t *testing.T) {
			rec := spaRequest(t, spaTestFS(), http.MethodGet, target)

			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}

			if strings.Contains(rec.Body.String(), "package web") {
				t.Error("handler served a file from outside the embedded tree")
			}
		})
	}
}

func TestSPAHeadSendsHeadersWithoutBody(t *testing.T) {
	rec := spaRequest(t, spaTestFS(), http.MethodHead, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got == "" {
		t.Error("Content-Type is missing")
	}

	if got := rec.Body.Len(); got != 0 {
		t.Errorf("body length = %d, want 0", got)
	}
}

func TestSPAOtherMethodsAreNotAllowed(t *testing.T) {
	methods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			rec := spaRequest(t, spaTestFS(), method, "/")

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
			}

			if got, want := rec.Header().Get("Allow"), "GET, HEAD"; got != want {
				t.Errorf("Allow = %q, want %q", got, want)
			}
		})
	}
}

// The cases an empty or stale build produces, checked against the real embedded
// dist. Reading embedded data is compile-time work, so this breaks no rule.
func TestEmbeddedSPABuildIsServable(t *testing.T) {
	fsys, err := fs.Sub(spaDist, spaDistDir)
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}

	if _, err := fs.Stat(fsys, spaIndexFile); err != nil {
		t.Fatalf("embedded dist has no %s; run `make ui-build` first: %v", spaIndexFile, err)
	}

	rec := spaRequest(t, fsys, http.MethodGet, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d, want %d", rec.Code, http.StatusOK)
	}

	// Every referenced asset must exist, which catches a stale index leftover
	// from an older build without hardcoding the hashed names.
	for _, ref := range spaAssetRefs(t, rec.Body.String()) {
		if _, err := fs.Stat(fsys, ref); err != nil {
			t.Errorf("index references %q, which is absent from the embedded build: %v", ref, err)
		}
	}
}

var spaAssetRefPattern = regexp.MustCompile(`(?:src|href)="([^"]+)"`)

// spaAssetRefs returns the local files referenced by an HTML document,
// normalising the leading "./" Vite emits because of `base: "./"`.
func spaAssetRefs(t *testing.T, html string) []string {
	t.Helper()

	var refs []string

	for _, match := range spaAssetRefPattern.FindAllStringSubmatch(html, -1) {
		ref := match[1]

		// Skip absolute URLs and data: URIs, which are not served from dist.
		if strings.Contains(ref, ":") || strings.HasPrefix(ref, "#") {
			continue
		}

		refs = append(refs, strings.TrimPrefix(ref, "./"))
	}

	if len(refs) == 0 {
		t.Fatal("index.html references no assets, which means the build is broken")
	}

	return refs
}

// The flag has to reach the entry document, because that is the only thing the
// UI reads before it renders. There is no endpoint for it by design.
func TestSPAInjectsAuthFlag(t *testing.T) {
	tests := []struct {
		name         string
		authRequired bool
		want         string
	}{
		{name: "authentication off", authRequired: false, want: "window.requireAuth=false"},
		{name: "authentication on", authRequired: true, want: "window.requireAuth=true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := spaRequestAuth(t, spaTestFS(), http.MethodGet, "/", tt.authRequired)

			body := rec.Body.String()

			if !strings.Contains(body, tt.want) {
				t.Errorf("index.html does not carry %q: %s", tt.want, body)
			}

			// It must run before the entry module, which is in the body.
			if strings.Index(body, "window.requireAuth") > strings.Index(body, "assets/index-CT2NxLkI.js") {
				t.Error("the flag is injected after the entry module, so it would be read too late")
			}
		})
	}
}

// Assets are content-hashed and cached forever, so rewriting them would make the
// served bytes disagree with the name the browser cached.
func TestSPAAuthFlagIsNotInjectedIntoAssets(t *testing.T) {
	rec := spaRequestAuth(t, spaTestFS(), http.MethodGet, "/assets/index-CT2NxLkI.js", true)

	if got := rec.Body.String(); strings.Contains(got, "requireAuth") {
		t.Errorf("asset was rewritten: %q", got)
	}

	if got, want := rec.Body.String(), "console.log(1)"; got != want {
		t.Errorf("body = %q, want the original asset %q", got, want)
	}
}

// Content-Length is computed after injection, so the browser is not told the
// original size of the file.
func TestSPAInjectedIndexHasMatchingContentLength(t *testing.T) {
	rec := spaRequestAuth(t, spaTestFS(), http.MethodGet, "/", true)

	if got, want := rec.Header().Get("Content-Length"), strconv.Itoa(rec.Body.Len()); got != want {
		t.Errorf("Content-Length = %q, body length = %s", got, want)
	}
}

// A document without a head is malformed, but the flag still has to be there:
// defaulting to "no auth" would silently unlock a deployment that needs a token.
func TestSPAInjectsAuthFlagWithoutHead(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`<body><div id="root"></div></body>`)},
	}

	rec := spaRequestAuth(t, fsys, http.MethodGet, "/", true)

	if got := rec.Body.String(); !strings.Contains(got, "window.requireAuth=true") {
		t.Errorf("flag missing from a headless document: %q", got)
	}
}
