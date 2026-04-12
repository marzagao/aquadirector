BINARY=aquadirector
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS=-ldflags "-X github.com/marzagao/aquadirector/cmd.Version=$(VERSION) -X github.com/marzagao/aquadirector/cmd.Commit=$(COMMIT)"

.PHONY: build test lint install clean

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test ./... -v

lint:
	go vet ./...

install:
	go install $(LDFLAGS) .

clean:
	rm -f $(BINARY)
