# ADR 002: Collections as Directories

## Status
Accepted

## Context
The user maintains multiple logically separate sets of journal entries (e.g., personal, work). The system needs to support an arbitrary number of these groupings without hardcoding them.

## Decision
Each collection is a directory on disk. A YAML config file (`~/.config/jot/config.yaml`) maps collection names to directory paths:

```yaml
collections:
  personal: ~/journals/personal
  work: ~/journals/work
```

- **Writes** target a single named collection (the user specifies which).
- **Reads** aggregate entries across all collections by default, with an optional `--collection` filter.

## Consequences
- Adding a new collection is a one-line config change — no code or schema changes needed.
- Each collection is fully self-contained on disk. Collections can be backed up, moved, or deleted independently.
- The read layer must scan all collection directories on every query. This is fine at personal journal scale.
- Collection names are user-defined and decoupled from directory names, so paths can be reorganized without breaking CLI usage.
