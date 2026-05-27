# tracker

A lightweight file-backed habit tracker.

Habits are YAML files under `habits/active/` or `habits/archived/`.
Events are monthly JSONL files under `events/`.

## Quick start

```sh
export TRACKER_ROOT="$HOME/habits"
tracker init
tracker habit add exercise --name "Exercise" --weekly 3
tracker log exercise
tracker review
tracker today
tracker week
```

By default, `tracker` uses the current working directory as the tracker root.
Set `TRACKER_ROOT` to use the same tracker from anywhere, or pass `--root` for a
single command:

```sh
tracker --root "$HOME/habits" week
```

All commands except `init` expect the tracker root to already contain
`habits/active/`, `habits/archived/`, and `events/`. If those directories are
missing, the command fails and asks you to run `tracker init`.

## Layout

```text
habits/
  active/
    exercise.yaml
  archived/
events/
  2026-05.jsonl
```

## Habit file

```yaml
id: exercise
name: Exercise
description: Any intentional workout, run, or strength session.
timezone: America/Sao_Paulo
created_at: 2026-05-08
schedule:
  type: weekly
  target: 3
```

Supported schedule types are `daily`, `weekly`, and `monthly`.
Weekly habits can optionally include fixed days:

```yaml
schedule:
  type: weekly
  target: 1
  days: [mon, wed, fri]
```

Create an editable habit file template:

```sh
tracker habit template drink_water water
```

If the file name has no directory, the template is written under
`habits/active/`. If it has no `.yaml` or `.yml` suffix, `.yaml` is added.

## Event record

```json
{"id":"20260508T224200000000000Z-exercise","habit_id":"exercise","date":"2026-05-08","created_at":"2026-05-08T19:42:00-03:00","note":"run"}
```

Older event lines with `done_at` are still readable, but new events use `date`.

## Review

`tracker review` opens an editor with a markdown checklist of habits due for the
review date. The date defaults to yesterday.

```sh
tracker review
tracker review --date 2026-05-08
```

The checklist looks like:

```md
- [ ] exercise Exercise
- [ ] read Read
```

Mark completed habits with `x`, save, and close the editor:

```md
- [x] exercise Exercise
```

Review includes daily habits, monthly habits, weekly habits scheduled for that
weekday, and weekly habits without fixed days.
