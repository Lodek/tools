# ADR 001: Flat File Storage Over a Database

## Status
Accepted

## Context
Jot needs persistent storage for journal entries. The two main options considered were a database (SQLite, Postgres) and plain markdown files on disk.

The system must support:
- Multiple independent collections of entries
- Easy ingestion of existing markdown journal files
- Graceful retirement — if the tool is abandoned, the data should remain useful without it

## Decision
Use plain markdown files as the storage layer. Each file covers one calendar month and is named `YYYY-MM.md`. Entries within a file use setext-style markdown headers (`===` or `---` underlines) as delimiters.

## Consequences
- **No migrations**: adding fields or changing behavior doesn't require schema changes.
- **Portability**: journal data is human-readable and editable with any text editor, independent of jot.
- **No query index**: filtering requires scanning all files. This is acceptable for personal journal volumes (hundreds to low thousands of entries).
- **Append-only writes are simple**: new entries are appended to the appropriate month file.
- **Conflict resolution is manual**: if the same file is edited outside jot, there's no merge logic. This is an accepted trade-off for simplicity.
