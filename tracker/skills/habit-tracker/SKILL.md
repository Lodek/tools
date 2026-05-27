---
name: habit-tracker
description: Use when working with this repository's file-backed habit tracker, including initializing a tracker root, adding or archiving YAML habit definitions, logging JSONL habit events, querying reports, or editing tracker data files directly.
---

# Habit Tracker

Use this skill for the lightweight Go habit tracker in this repository.

## Root Selection

The tracker root is where data files live:

```text
habits/active/
habits/archived/
events/
```

Root precedence:

1. `tracker --root <dir> ...`
2. `TRACKER_ROOT`
3. current working directory

All commands except `tracker init` fail fast if the root is missing `habits/active/`, `habits/archived/`, or `events/`.

## Prefer The CLI

Use the CLI for normal operations:

```sh
tracker init
tracker habit add exercise --name "Exercise" --weekly 3 --days mon,wed,fri
tracker habit template drink_water water
tracker log exercise --date 2026-05-08 --note "run"
tracker review --date 2026-05-08
tracker habit archive exercise
tracker events --habit exercise --from 2026-05-01 --to 2026-05-31
tracker today
tracker week
tracker month
```

When running from source, use:

```sh
go run . <command>
```

If the local Go cache is read-only, use the project Makefile or set:

```sh
GOCACHE=/tmp/tracker-gocache
```

## Habit Files

Active habits live in `habits/active/<id>.yaml`. Archived habits live in `habits/archived/<id>.yaml`.

Example:

```yaml
id: exercise
name: Exercise
description: Any intentional workout, run, or strength session.
timezone: America/Sao_Paulo
created_at: 2026-05-08
schedule:
  type: weekly
  target: 3
  days: [mon, wed, fri]
```

Rules:

- Habit IDs should contain only lowercase letters, numbers, dashes, and underscores.
- Supported schedule types are `daily`, `weekly`, and `monthly`.
- `target` must be at least `1`.
- For archived habits, move the file to `habits/archived/`; do not delete it.
- Use `tracker habit template <id> <file>` to create a default editable habit file.

## Event Files

Events are monthly JSONL files under `events/YYYY-MM.jsonl`.

Example line:

```json
{"id":"20260508T224200000000000Z-exercise","habit_id":"exercise","date":"2026-05-08","created_at":"2026-05-08T19:42:00-03:00","note":"run"}
```

Rules:

- Append one JSON object per line.
- `date` is the completion date in `YYYY-MM-DD`.
- `created_at` is when the event was recorded.
- Keep `habit_id` stable; historical events should keep linking to archived habit files.
- Prefer the CLI for event creation so IDs and monthly file placement stay consistent.
- Older logs using `done_at` are readable, but new events should use `date`.

## Review

Use `tracker review` to log due habits in one editor pass. The date defaults to yesterday and can be set with `--date YYYY-MM-DD`.

Review includes daily habits, monthly habits, weekly habits scheduled for that weekday, and weekly habits without fixed days. The buffer is a markdown checklist; checked lines `- [x] <habit-id> ...` are logged for the review date after the editor closes.

## Validation

After changing code or skill-relevant behavior, run:

```sh
make test
```

For a full local check:

```sh
make check
```
