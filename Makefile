BINARY_NAME ?= coldwire-server
BIN_DIR ?= bin

# Default OS/ARCH for local builds
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Windows executable extension when building for windows
ifeq ($(GOOS),windows)
  EXT := .exe
else
  EXT :=
endif

.PHONY: build clean

# If CI=true is set, only create the suffixed binaries (avoid creating the plain binary)
build:
	@mkdir -p $(BIN_DIR)
ifeq ($(CI),true)
	@echo "CI build: producing suffixed binary only"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)$(EXT) ./cmd/server
else
	@echo "Local build: producing suffixed binary and plain binary"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)$(EXT) ./cmd/server
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/$(BINARY_NAME)$(EXT) ./cmd/server
endif

clean:
	rm -rf $(BIN_DIR)
