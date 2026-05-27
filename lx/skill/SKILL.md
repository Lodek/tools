---
description: "Interact with lx — a CLI for managing flat markdown list files (todos, reference lists). Use this skill to read, create, add items, and clean up lists."
---

# lx — List Management Skill

You have access to `lx`, a CLI tool for managing the user's markdown list files. Lists live as flat `.md` files in the directory specified by `$LX_DIR`. Each file is a "collection" containing sublists grouped under headers.

## Key Concepts

- **Collection**: a markdown file (e.g. `todo.md` → collection name is `todo`)
- **Sublist**: a `# header` section within a collection. Bold (`# **name**`) = active, plain (`# name`) = parked/inactive.
- **Items**: lines under a sublist header. Either actionable (`- [ ] task`) or reference (`- item`).
- **Tags**: inline `<tag>` markers on items (e.g. `<this-week>`, `<this-month>`)
- **Done items**: `- [x] task` — completed actionable items awaiting cleanup

## Commands

### Discover what lists exist

```bash
lx list
```

Prints each collection name and its description (from frontmatter).

### Read a list

```bash
lx get <collection>                # full collection
lx get <collection>.<sublist>      # specific sublist only
lx get <collection> --active       # only active (bold) sublists
lx get <collection> -t <tag>       # only items with this tag
lx get --all                       # everything across all collections
lx get --all -t <tag>              # everything with a specific tag
```

Filters combine: `lx get todo.eurotrip -t this-week` shows only items in the eurotrip sublist tagged `<this-week>`.

### Add an item

```bash
lx add <collection> <sublist> "<item text>"
lx add <collection> <sublist> "<item text>" --tag <tag>
```

Appends a `- [ ] item` to the specified sublist. Creates the sublist (as inactive) if it doesn't exist. The `--tag` flag adds an inline `<tag>` to the item.

### Create a new collection

```bash
lx create <name>
lx create <name> -d "<description>"
```

Creates a new empty `.md` file. Errors if it already exists.

### Clean up completed items

```bash
lx done --dry-run    # preview what would be removed
lx done              # remove [x] items across all collections, log to lx.json
```

Iterates over all collections, finds all `- [x]` items, logs them as JSONL records to `$LX_DIR/lx.json`, and removes them from the source files.

### Edit in $EDITOR

```bash
lx edit <collection>
```

Opens the raw markdown file in the user's editor. Use this when bulk edits are needed.

## Typical Agent Workflows

**Session start — surface priorities:**
```bash
lx get todo -t this-week
```

**Capture a new task:**
```bash
lx add todo house "Fix bathroom drain" --tag this-week
```

**Discover available lists:**
```bash
lx list
```

**Weekly review — see all active commitments:**
```bash
lx get todo --active
```

**End of session — clean up done items:**
```bash
lx done
```

**Check what's in a specific area:**
```bash
lx get todo.eurotrip
```

## Important Notes

- Always use `lx list` first if you're unsure what collections exist.
- The `--tag` value on the CLI does NOT include angle brackets — use `--tag this-week`, not `--tag "<this-week>"`.
- Items added with `lx add` are always actionable (`- [ ]`). For reference items, use `lx edit`.
- The done log (`lx.json`) is append-only JSONL. Each record: `{"list":"...","sublist":"...","date":"YYYY-MM-DD","entry":"..."}`.
- Files are plain markdown — the user may edit them by hand at any time. Don't assume your last read is still current.
