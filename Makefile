BINARY_NAME = coldwire-server
BIN_DIR = bin

# Default OS/ARCH for local builds
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: build clean

build:
	@mkdir -p $(BIN_DIR)

	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -trimpath -o $(BIN_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) ./cmd/server


clean:
	rm -rf $(BIN_DIR)
