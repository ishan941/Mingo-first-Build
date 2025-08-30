BIN_DIR := bin

.PHONY: all build test clean lex repl vmrepl run

all: build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/lex ./cmd/lex
	go build -o $(BIN_DIR)/repl ./cmd/repl
	go build -o $(BIN_DIR)/run ./cmd/run
	go build -o $(BIN_DIR)/vmrepl ./cmd/vmrepl

test:
	go test ./...

lex: build
	@if [ -z "$(FILE)" ]; then echo "Usage: make lex FILE=path/to/file.mg"; exit 2; fi
	$(BIN_DIR)/lex $(FILE)

repl: build
	$(BIN_DIR)/repl

vmrepl: build
	$(BIN_DIR)/vmrepl

run: build
	@if [ -z "$(FILE)" ]; then echo "Usage: make run FILE=path/to/file.mg"; exit 2; fi
	cat $(FILE) | $(BIN_DIR)/run

clean:
	rm -rf $(BIN_DIR)
