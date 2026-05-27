BINARY := lx
BUILD_DIR := build

SKILL_DIR := $(HOME)/.claude/skills/lx

.PHONY: build install install-skill test clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) .

install:
	go install .

test:
	go test ./...

install-skill:
	mkdir -p $(SKILL_DIR)
	cp -r skill/* $(SKILL_DIR)/

clean:
	rm -rf $(BUILD_DIR)
