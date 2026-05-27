# jot

A CLI tool for reading and writing markdown journal entries across multiple collections.

## Install

```sh
go build -o jot ./cmd/jot/
# move to somewhere on your PATH
mv jot ~/.local/bin/
```

## Configuration

Config defaults to `~/.jot.yaml`. Override with the `JOT_CONFIG` environment variable:

```sh
export JOT_CONFIG=/path/to/config.yaml
```

Config format:

```yaml
collections:
  personal: ~/journals/personal
  work: ~/journals/work
```

Each collection maps a name to a directory of markdown files.

## Usage

### Read entries

```sh
# all entries across all collections
jot read

# filter by date range
jot read --from 2024-03-01 --to 2024-03-31

# filter by month
jot read --month 2024-03

# filter by ISO week number
jot read --week 11

# filter by tag
jot read --tag climbing

# filter by collection
jot read --collection personal
```

Filters can be combined.

### Write a new entry

```sh
# opens $EDITOR with a dated template, saves to the collection
jot write personal

# with tags
jot write personal --tag reflection --tag climbing
```

### Ingest existing files

```sh
# parse markdown files and merge entries into a collection by month
jot ingest personal old-journal.md another-file.md
```

## Entry format

Entries use setext-style markdown headers with optional tags:

```
2024-03-11 [reflection] [climbing]
===================================

Entry body goes here.
```

Files are organized by month (`YYYY-MM.md`), one file per month per collection.

## Architecture

See [docs/adr/](docs/adr/) for decision records.
