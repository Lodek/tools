# ADR 004: Go as the Implementation Language

## Status
Accepted

## Context
The tool needs to be:
- A CLI binary with no runtime dependencies
- Typed enough to catch structural bugs during development
- Simple to build, distribute, and maintain long-term
- Potentially extensible to a gRPC service later

Candidates considered: Python, Go, Rust, Haskell.

## Decision
Use Go.

## Rationale
- **Single static binary**: no virtualenv, no interpreter, no dynamic linking. `go build` produces a self-contained executable.
- **Type safety without ceremony**: the domain (dates, tags, file I/O, filtering) benefits from types but doesn't need Rust's ownership model or Haskell's type system.
- **Stdlib coverage**: `os`, `bufio`, `time`, `regexp`, `os/exec`, `path/filepath` handle nearly all of the tool's needs.
- **gRPC readiness**: Go has first-class gRPC/protobuf support via `google.golang.org/grpc`, making a future service layer straightforward to add.
- **Fast compile/edit cycle**: relevant for iterating on a personal tool.

### Why not the others
- **Python**: fast to prototype but lack of types would hurt as query/filter logic grows. Packaging and distribution is painful.
- **Rust**: borrow checker overhead for a tool with no concurrency or shared-state complexity. Slower iteration.
- **Haskell**: build tooling (cabal/stack) and GHC runtime add operational friction — the opposite of what this project optimizes for.

## Consequences
- The team (or future agents) need Go familiarity to contribute.
- Error handling is verbose but explicit — no hidden exceptions.
- Dependency footprint is minimal (cobra for CLI, yaml.v3 for config).
