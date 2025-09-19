BINARY_NAME = coldwire-server
BIN_DIR = bin

# Default OS/ARCH for local builds
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: build clean

build:
	@mkdir -p $(BIN_DIR)

	# Build for CI: include OS/ARCH suffix
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) ./cmd/server

	# Build for local use: plain binary
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/server

clean:
	rm -rf $(BIN_DIR)
