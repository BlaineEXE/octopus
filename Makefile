GO_IMPORT_ROOT ?= github.com/BlaineEXE/octopus
GO_BUILD_TARGET := $(GO_IMPORT_ROOT)/cmd/octopus
OUTPUT_DIR ?= _output

BUILDFLAGS ?= -buildmode pie
LDFLAGS ?=

# Version is automatically set by getting tag info from git; can be set manually if desired
VERSION ?= $(shell git describe --dirty --always --tags | sed 's/-/./2' | sed 's/-/./2' )

# inject the version number into the golang version package using the -X linker flag
LDFLAGS += -X $(GO_IMPORT_ROOT)/internal/version.Version=$(VERSION)

ALL_BUILDFLAGS := $(BUILDFLAGS) -ldflags '$(LDFLAGS)'

dep:
	dep ensure

build: dep
	go build $(ALL_BUILDFLAGS) -o $(OUTPUT_DIR)/octopus $(GO_BUILD_TARGET)
	@ echo "Binary size: $$(ls -sh $(OUTPUT_DIR)/octopus | cut -d' ' -f 1)"

test: dep
	go test -cover ./cmd/... ./internal/...

install: dep
	go install $(ALL_BUILDFLAGS) $(GO_BUILD_TARGET)

clean:
	rm -rf $(OUTPUT_DIR)
	rm -rf vendor/
