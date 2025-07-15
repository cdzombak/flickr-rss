BINARY = flickr-rss
GOBUILD = go build
GOCLEAN = go clean
GOTEST = go test
GOGET = go get
GOMOD = go mod

.PHONY: all build clean test deps help

all: build

build:
	$(GOBUILD) -o $(BINARY) -v

clean:
	$(GOCLEAN)
	rm -f $(BINARY)

test:
	$(GOTEST) -v ./...

deps:
	$(GOMOD) download
	$(GOMOD) tidy

install: build
	cp $(BINARY) /usr/local/bin/

help:
	@echo "Available targets:"
	@echo "  build    - Build the binary"
	@echo "  clean    - Clean build artifacts"
	@echo "  test     - Run tests"
	@echo "  deps     - Download and organize dependencies"
	@echo "  install  - Install binary to /usr/local/bin"
	@echo "  help     - Show this help message"