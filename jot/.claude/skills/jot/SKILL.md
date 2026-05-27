---
name: jot
description: Use when the user asks about their journal, wants to read/query journal entries, write a new entry, or ingest journal files. Also use when you need to understand how the jot CLI works.
user-invocable: false
---

# jot — Journal CLI

`jot` is a CLI tool for reading and writing markdown journal entries. Entries are stored as flat markdown files organized by month (`YYYY-MM.md`) across multiple directories called "collections".

## Configuration

Config defaults to `~/.jot.yaml`. Override with the `JOT_CONFIG` environment variable:

```yaml
default_collection: personal
collections:
  personal: ~/journals/personal
  work: ~/journals/work
```

- `default_collection`: used when no collection is specified in commands
- `collections`: maps names to directories of markdown files

Initialize a config with `jot init`.

## Entry Format

Entries use setext-style markdown headers with optional tags:

```
2024-03-11 [tag1] [tag2]
=========================

Entry body text here.
```

- Date is always `YYYY-MM-DD`
- Tags are optional, in square brackets after the date
- The underline uses `=` or `-`, at least 3 characters
- Multiple entries can share the same date
- Files are named `YYYY-MM.md`, one per month per collection

## Commands

### Read entries

```sh
jot read                              # all entries, all collections
jot read --from 2024-03-01            # from a start date
jot read --to 2024-03-31              # up to an end date
jot read --month 2024-03              # specific month
jot read --week 11                    # ISO week number
jot read --tag climbing               # filter by tag
jot read --collection personal        # filter by collection
```

Flags can be combined. Output format matches the entry format (setext headers with `=` underlines).

### Write a new entry

```sh
jot new                               # uses default collection
jot new personal                      # explicit collection
jot new --tag reflection --tag work   # with tags
```

Opens `$EDITOR` with a dated template. Saves to the correct `YYYY-MM.md` file on exit. Aborts if the entry is left empty.

### Add an entry from stdin

```sh
echo "Entry body here." | jot add                        # default collection, today's date
echo "Entry body." | jot add work                        # explicit collection
echo "Entry body." | jot add --date 2024-12-25           # specific date
echo "Entry body." | jot add --tag reflection --tag work # with tags
```

Reads the entry body from stdin, wraps it in the correct header format, and appends to the right month file. This is the preferred way for agents and scripts to create entries.

### Ingest existing files

```sh
jot ingest file1.md file2.md          # into default collection
jot ingest work file1.md file2.md     # into explicit collection
```

Parses markdown files and merges entries into the target collection, splitting them into the correct month files.

## Important Notes for Agents

- **Use `jot add`** to create entries programmatically — pipe the body text to stdin.
- **Do not call `jot new`** — it opens an interactive editor and will hang.
- `jot read` is safe to call non-interactively and returns entries to stdout.
- Set `JOT_CONFIG` to override the default config path (`~/.jot.yaml`).
