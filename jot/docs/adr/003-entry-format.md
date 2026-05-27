# ADR 003: Entry Format — Setext Headers with Optional Tags

## Status
Accepted

## Context
Journal entries need a delimiter format that is:
- Valid markdown (entries are readable without jot)
- Parseable by the tool
- Compatible with existing journal files the user already has

## Decision
Each entry starts with a setext-style markdown header:

```
2024-03-11 [tag1] [tag2]
=========================
```

The format is:
- **Line 1**: ISO date (`YYYY-MM-DD`), optionally followed by space-separated tags in square brackets (`[tag]`)
- **Line 2**: three or more `=` or `-` characters (setext underline)
- **Body**: everything after the underline until the next entry header or end of file

### Parsing rules
- The date regex is `^\d{4}-\d{2}-\d{2}` — strict ISO format only.
- Tags are extracted via `\[([^\]]+)\]` from the remainder of the header line.
- Multiple entries can share the same date within a file.
- A file can contain entries from any dates, though by convention files are grouped by month.

## Consequences
- Existing journal files in this format can be ingested without modification.
- The format is a subset of standard markdown — any markdown viewer renders it correctly.
- Tags are lightweight and optional. No predefined tag vocabulary is enforced.
- The parser is simple (regex + line scanning) with no need for a full markdown AST.
- Limitation: entry bodies cannot contain lines that look like a date header followed by a setext underline. This is unlikely in practice.
