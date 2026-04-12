BINARY=aquadirector
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS=-ldflags "-X github.com/marzagao/aquadirector/cmd.Version=$(VERSION) -X github.com/marzagao/aquadirector/cmd.Commit=$(COMMIT)"

GOBIN ?= $(shell go env GOPATH)/bin

GOCYCLO     := $(GOBIN)/gocyclo
INEFFASSIGN := $(GOBIN)/ineffassign
MISSPELL    := $(GOBIN)/misspell

.PHONY: build test lint check tools install clean

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test ./... -v

lint:
	go vet ./...

tools:
	@command -v gocyclo     >/dev/null 2>&1 || go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@command -v ineffassign >/dev/null 2>&1 || go install github.com/gordonklaus/ineffassign@latest
	@command -v misspell    >/dev/null 2>&1 || go install github.com/client9/misspell/cmd/misspell@latest

check: tools
	@echo "==> gofmt"
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "$$out"; echo "run: gofmt -w ."; exit 1; fi
	@echo "==> go vet"
	@go vet ./...
	@echo "==> gocyclo (threshold 15)"
	@$(GOCYCLO) -over 15 .
	@echo "==> ineffassign"
	@$(INEFFASSIGN) ./...
	@echo "==> misspell"
	@$(MISSPELL) -error .
	@echo "==> license"
	@test -f LICENSE || (echo "LICENSE file missing" && exit 1)
	@echo "all checks passed"

install:
	go install $(LDFLAGS) .

clean:
	rm -f $(BINARY)
