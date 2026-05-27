.PHONY: build check fmt test tidy install

BINARY ?= tracker
GOCACHE ?= /tmp/tracker-gocache

build:
	mkdir -p bin
	GOCACHE=$(GOCACHE) go build -o bin/$(BINARY) .

check: fmt tidy test

fmt:
	gofmt -w main.go cmd/*.go internal/tracker/*.go

test:
	GOCACHE=$(GOCACHE) go test ./...

tidy:
	GOCACHE=$(GOCACHE) go mod tidy

install:
	go install .