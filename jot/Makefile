.PHONY: build
build:
	go build -o build/jot cmd/jot/main.go

.PHONY: install
install:
	go install ./cmd/jot

.PHONY: install-skill
install-skill:
	cp -r skills ~/.claude/skills/jot
