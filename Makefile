ifndef GOOS
# Workaround for Travis CI: load go environment variables
export GOOS := $(shell go env GOOS)
export GOARCH := $(shell go env GOARCH)
endif

# Entering/Leaving directory messages are annoying and not very useful
MAKEFLAGS += --no-print-directory

GO_IMPORT_ROOT ?= github.com/BlaineEXE/octopus
GO_BUILD_TARGET := $(GO_IMPORT_ROOT)/cmd/octopus
OUTPUT_DIR ?= _output

BUILDFLAGS ?= -buildmode pie

# Version is automatically set by getting tag info from git; can be set manually if desired
VERSION ?= $(shell git describe --dirty --always --tags | sed 's/-/./2' | sed 's/-/./2' )

# inject the version number into the golang version package using the -X linker flag
LDFLAGS += -X $(GO_IMPORT_ROOT)/internal/version.Version=$(VERSION)

ALL_BUILDFLAGS := $(BUILDFLAGS) -ldflags '$(LDFLAGS)'

# Allow setting this to not-false to test binaries built by crossbuild instead of regular build
# e.g., for CI testing so binaries deployed after CI are more sure to be working.
TEST_CROSSBUILD_BINARIES ?= false
XGO_FLAGS ?=

BINARY_NAME ?= $(OUTPUT_DIR)/octopus-static-$(VERSION)-$(GOOS)-$(GOARCH)

.PHONY: vendor
vendor:
	@ dep ensure

build: vendor
	go build $(ALL_BUILDFLAGS) -o $(BINARY_NAME) $(GO_BUILD_TARGET)
	@ echo "Binary size: $$(ls -sh $(BINARY_NAME) | cut -d' ' -f 1)"

test: vendor
	go test -cover ./cmd/... ./internal/...

test.integration: vendor build
	@ mkdir -p test/_output
	cp $(BINARY_NAME) test/_output/octopus
	@ $(MAKE) --directory test all

test.smoke: vendor build
	@ mkdir -p test/_output
	cp $(BINARY_NAME) test/_output/octopus
	@ $(MAKE) --directory test smoke

install: vendor
	go install $(ALL_BUILDFLAGS) $(GO_BUILD_TARGET)

completion:
	@ # Currently only generates a basic completion shell file and prints to stdout
	@ go run cmd/octopus/bashcompletion/generate.go

clean:
	rm -rf $(OUTPUT_DIR)
	rm -rf vendor/
	@ $(MAKE) --directory test teardown

clean.all: clean
	@ rm -rf _tools/
