.PHONY: build test clean

BINARY_NAME = git-time-machine
BINARY_PATH = ./bin/$(BINARY_NAME)

build: clean
	@mkdir -p bin
	go build -o $(BINARY_PATH) ./cmd/main.go

test:
	go test ./...

clean:
	rm -rf bin