# pm — project metadata manager

`pm` manages project metadata (notes, journals) separately from code repositories. Projects are organized into **collections** — directories that group related projects (e.g., personal, work). Each project lives as a subdirectory under a collection with a `README.md` containing structured frontmatter metadata.

When a project is linked to a working directory, `pm` symlinks project files using a `.pm.md` suffix so they can be globally gitignored.

## Install

Requires Python 3.10+.

```sh
python -m venv .venv
source .venv/bin/activate
pip install -e .
```

Or with pipx for global availability:

```sh
pipx install -e ~/src/pm
```

### Shell integration

`pm jump` needs a shell wrapper to change your working directory. Add this to `.zshrc`:

```zsh
if command -v pm &>/dev/null; then
  eval "$(pm setup_shell)"
fi
```

### Git ignore

Add `*.pm.md` to your global gitignore:

```sh
echo '*.pm.md' >> ~/.config/git/ignore
```

## Usage

### Register a collection

Before creating projects, register one or more collection root directories:

```sh
pm collection add personal ~/projects/personal
pm collection add work ~/projects/work
pm collection list
pm collection remove old-collection
```

### Create a project

```sh
pm init -n my-app -c personal
pm init -n client-portal -c work -t "react frontend" -s ongoing
```

This creates a subdirectory under the collection root with:
- `README.md` — frontmatter with name, status, start-date, tags
- `notes.pm.md` — freeform notes and brainstorming
- `journal.pm.md` — timestamped journal entries

### Link a workspace

```sh
# Link current directory
pm link my-app

# Link a specific directory
pm link my-app -p ~/src/my-app
```

Creates symlinks to `notes.pm.md` and `journal.pm.md` in the workspace and records the workspace path in the project's frontmatter.

### Edit project files

```sh
pm edit my-app notes     # default
pm edit my-app journal
```

Opens the file in `$EDITOR` (falls back to `vim`).

### Jump to a project directory

```sh
pm jump meta my-app    # cd to project metadata directory
pm jump work my-app    # cd to linked workspace
```

Requires the shell integration described above.

### List projects

```sh
pm list                    # all non-done projects
pm list --all              # include done projects
pm list -t python          # filter by tag
pm list -t "python react"  # multiple tags (OR logic)
pm list -c work            # filter by collection
```

## Project README frontmatter

Each project's `README.md` uses YAML-like frontmatter as the source of truth:

```markdown
---
name: my-app
status: active
start-date: 2026-04-28
tags: python cli
workspace: /home/user/src/my-app
---

Project description and scope go here.
```

**Status values:** `active`, `ongoing`, `archived`, `done`

## File locations

| What | Where |
|------|-------|
| Config | `~/.pm/config.json` |
| Project metadata | `<collection-root>/<project>/README.md` |
| Project files | `notes.pm.md`, `journal.pm.md` in the project directory |
