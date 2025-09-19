BINARY_NAME = coldwire-server
BIN_DIR = bin

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: build clean

build:
	mkdir -p $(BIN_DIR)

	# this for ci
    GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) ./cmd/server

	# this is for local builds
    GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/server

clean:
	rm -rf $(BIN_DIR)
