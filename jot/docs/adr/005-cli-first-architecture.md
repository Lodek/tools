# ADR 005: CLI-First Architecture with Service Layer Planned

## Status
Accepted

## Context
The tool serves two audiences over time:
1. **Now**: the user interacting via terminal
2. **Later**: AI agents or other programs interacting programmatically

A gRPC service layer is planned for the future but not yet needed.

## Decision
Build as a CLI tool first using cobra for command structure. Keep the core logic in `internal/journal/` as a library, with the CLI in `cmd/jot/` as a thin wrapper.

```
cmd/jot/main.go        → CLI layer (cobra commands, flag parsing, output formatting)
internal/journal/      → core library (parsing, querying, writing — no I/O formatting)
internal/config/       → config loading
```

## Consequences
- The core library (`internal/journal`) has no dependency on cobra or CLI concerns. A future gRPC service would import the same library and wrap it differently.
- Adding a `cmd/jotd/` or `internal/server/` for gRPC later requires no refactoring of the journal package.
- The CLI is the only interface for now, keeping scope small and iteration fast.
- `internal/` prevents external imports, which is fine — this is a personal tool, not a library. If the journal package ever needs to be shared, it can be moved to `pkg/`.
