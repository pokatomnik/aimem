GO ?= go

BINARY     := aimem
BUILD_DIR  ?= build
CMD        ?= ./cmd/aimem

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

.PHONY: all build install test clean

all: build

build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build $(GO_BUILD_FLAGS) -ldflags="$(GO_LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(CMD)

install: build
	@mkdir -p $(INSTALL_DIR)
	install -m 0755 $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)

test:
	CGO_ENABLED=0 $(GO) test ./...

clean:
	rm -rf $(BUILD_DIR)
