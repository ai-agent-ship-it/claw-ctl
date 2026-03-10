BINARY_NAME=claw-ctl
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

.PHONY: build clean test run install

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)

test:
	go test ./...

run: build
	./$(BINARY_NAME)

install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

lint:
	golangci-lint run ./...

.PHONY: presets
presets: build
	./$(BINARY_NAME) presets
