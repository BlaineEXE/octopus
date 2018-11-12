
TOOLS_DIR := _tools

DEP_VERSION ?= v0.5.0
DEP := $(TOOLS_DIR)/dep-$(DEP_VERSION)
$(DEP):
	@ echo ' === downloading dep version $(DEP_VERSION) === '
	@ mkdir -p $(TOOLS_DIR)
	@ wget https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-linux-amd64 -O $(DEP)
	@ chmod +x $(DEP)


GO_VERSION ?= 1.11
GOROOT := $(TOOLS_DIR)/go-$(GO_VERSION)
export PATH := $(GOROOT)/bin:$(PATH)
GO := $(GOROOT)/bin/go
GO_TARFILE := go${GO_VERSION}.linux-amd64.tar.gz
$(GO):
	@ echo ' === downloading go version $(GO_VERSION) === '
	@ curl --location --create-dirs https://dl.google.com/go/$(GO_TARFILE) --output $(GO_TARFILE)
	@ tar -C $(TOOLS_DIR) -xzf $(GO_TARFILE)
	@ mv $(TOOLS_DIR)/go/ $(GOROOT)
	@ rm -f go${GO_VERSION}.linux-amd64.tar.gz
