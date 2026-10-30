GO ?= go
NPM ?= npm

BINARY     := aimem
BUILD_DIR  ?= build
CMD        ?= ./cmd/aimem

UI_DIR   := internal/controllers/web/ui
UI_DIST  := $(UI_DIR)/dist
# Reinstall only when the lockfile or manifest changes.
UI_STAMP := $(UI_DIR)/node_modules/.install-stamp

ifeq ($(shell $(GO) env GOOS),windows)
BINARY := $(BINARY).exe
endif

# Fall back to GOPATH/bin when GOBIN is not configured.
GOBIN := $(shell $(GO) env GOBIN)
ifeq ($(strip $(GOBIN)),)
GOBIN := $(shell $(GO) env GOPATH)/bin
endif

INSTALL_DIR ?= $(GOBIN)

# Strip paths, VCS metadata and debug information for the smallest binary.
GO_BUILD_FLAGS := -trimpath -buildvcs=false
GO_LDFLAGS     := -s -w -buildid=

# node_modules may contain Go code of its own (flatted ships a Go port), which
# is not ours to test. Expanded at recipe time, so dist/ exists by then.
GO_PACKAGES = $$($(GO) list ./... | grep -v /node_modules)

.PHONY: all build install test generate api-generate clean ui-install ui-build ui-test ui-lint ui-typecheck

all: build

# The UI is embedded into the binary with go:embed, so it must exist before the
# Go compiler runs. `ui-build` therefore comes first.
build: ui-build
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build $(GO_BUILD_FLAGS) -ldflags="$(GO_LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(CMD)

$(UI_STAMP): $(UI_DIR)/package.json $(UI_DIR)/package-lock.json
	cd $(UI_DIR) && $(NPM) ci
	@touch $(UI_STAMP)

ui-install: $(UI_STAMP)

ui-build: $(UI_STAMP)
	cd $(UI_DIR) && $(NPM) run build

ui-typecheck: $(UI_STAMP)
	cd $(UI_DIR) && $(NPM) run typecheck

ui-test: $(UI_STAMP)
	cd $(UI_DIR) && $(NPM) run test

ui-lint: $(UI_STAMP)
	cd $(UI_DIR) && $(NPM) run lint

install: build
	@mkdir -p $(INSTALL_DIR)
	install -m 0755 $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)

test: ui-build
	CGO_ENABLED=0 $(GO) test $(GO_PACKAGES)

# Regenerate code from //go:generate directives (currently ogen from OpenAPI).
generate:
	$(GO) generate ./...

# Regenerate both sides of the API contract: the Go server (ogen) and the UI
# client types (openapi-typescript). They are generated from the same
# openapi.yaml, so running one without the other leaves the client typed against
# stale shapes — a mismatch that only shows up as a tsc failure much later.
api-generate: generate
	cd $(UI_DIR) && $(NPM) run api:generate

clean:
	rm -rf $(BUILD_DIR) $(UI_DIST)
