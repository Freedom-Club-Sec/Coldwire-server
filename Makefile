BINARY_NAME = coldwire-server
BIN_DIR = bin

.PHONY: build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/server


clean:
	rm -rf $(BIN_DIR)
