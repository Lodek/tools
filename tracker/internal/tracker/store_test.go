package tracker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStoreAddLoadAndArchiveHabit(t *testing.T) {
	store := NewStore(t.TempDir())
	initTestStore(t, store)

	habit := Habit{
		ID:          "exercise",
		Name:        "Exercise",
		Description: "Workout",
		Timezone:    "America/Sao_Paulo",
		CreatedAt:   "2026-05-08",
		Schedule: Schedule{
			Type:   "weekly",
			Target: 3,
			Days:   []string{"mon", "wed", "fri"},
		},
	}
	if err := store.AddHabit(habit); err != nil {
		t.Fatalf("AddHabit() error = %v", err)
	}

	loaded, err := store.LoadHabit("exercise")
	if err != nil {
		t.Fatalf("LoadHabit() error = %v", err)
	}
	if loaded.ID != habit.ID || loaded.Name != habit.Name || loaded.Schedule.Target != habit.Schedule.Target {
		t.Fatalf("loaded habit mismatch: got %#v", loaded)
	}
	if got := DescribeSchedule(loaded.Schedule); got != "3/week on mon,wed,fri" {
		t.Fatalf("DescribeSchedule() = %q", got)
	}

	if err := store.AddHabit(habit); err == nil {
		t.Fatal("AddHabit() duplicate succeeded, want error")
	}

	if err := store.ArchiveHabit("exercise"); err != nil {
		t.Fatalf("ArchiveHabit() error = %v", err)
	}
	if _, err := store.LoadHabit("exercise"); err == nil || !strings.Contains(err.Error(), "archived") {
		t.Fatalf("LoadHabit() after archive error = %v, want archived error", err)
	}

	habits, err := store.LoadHabits(true)
	if err != nil {
		t.Fatalf("LoadHabits(true) error = %v", err)
	}
	if len(habits) != 1 || !habits[0].Archived {
		t.Fatalf("LoadHabits(true) = %#v, want one archived habit", habits)
	}
}

func TestStoreLogEventAndFilterByDate(t *testing.T) {
	root := t.TempDir()
	store := NewStore(root)
	initTestStore(t, store)
	addTestHabit(t, store, "read", Schedule{Type: "daily", Target: 1})

	event, err := store.LogEvent("read", "2026-05-08", "20 pages")
	if err != nil {
		t.Fatalf("LogEvent() error = %v", err)
	}
	if event.HabitID != "read" || event.Date != "2026-05-08" || event.Note != "20 pages" {
		t.Fatalf("logged event mismatch: got %#v", event)
	}

	logPath := filepath.Join(root, "events", "2026-05.jsonl")
	contents, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", logPath, err)
	}
	if !strings.Contains(string(contents), `"habit_id":"read"`) {
		t.Fatalf("event log does not contain habit_id: %s", contents)
	}
	if !strings.Contains(string(contents), `"date":"2026-05-08"`) {
		t.Fatalf("event log does not contain date field: %s", contents)
	}

	events, err := store.LoadEvents("2026-05-01", "2026-05-31")
	if err != nil {
		t.Fatalf("LoadEvents(in range) error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("LoadEvents(in range) length = %d, want 1", len(events))
	}

	events, err = store.LoadEvents("2026-06-01", "2026-06-30")
	if err != nil {
		t.Fatalf("LoadEvents(out of range) error = %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("LoadEvents(out of range) length = %d, want 0", len(events))
	}
}

func TestStoreLoadsLegacyDoneAtEvents(t *testing.T) {
	root := t.TempDir()
	store := NewStore(root)
	initTestStore(t, store)
	addTestHabit(t, store, "read", Schedule{Type: "daily", Target: 1})

	logPath := filepath.Join(root, "events", "2026-05.jsonl")
	if err := os.WriteFile(logPath, []byte(`{"id":"legacy","habit_id":"read","done_at":"2026-05-08","created_at":"2026-05-08T10:00:00Z"}`+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	events, err := store.LoadEvents("2026-05-01", "2026-05-31")
	if err != nil {
		t.Fatalf("LoadEvents() error = %v", err)
	}
	if len(events) != 1 || events[0].Date != "2026-05-08" {
		t.Fatalf("LoadEvents() = %#v, want legacy done_at mapped to Date", events)
	}
}

func TestStoreCreateHabitTemplate(t *testing.T) {
	root := t.TempDir()
	store := NewStore(root)
	initTestStore(t, store)

	path, err := store.CreateHabitTemplate("drink_water", "water")
	if err != nil {
		t.Fatalf("CreateHabitTemplate() error = %v", err)
	}
	wantPath := filepath.Join(root, "habits", "active", "water.yaml")
	if path != wantPath {
		t.Fatalf("CreateHabitTemplate() path = %q, want %q", path, wantPath)
	}

	loaded, err := store.readHabit(path, false)
	if err != nil {
		t.Fatalf("readHabit() error = %v", err)
	}
	if loaded.ID != "drink_water" || loaded.Name != "Drink Water" || loaded.Schedule.Type != "daily" || loaded.Schedule.Target != 1 {
		t.Fatalf("template habit = %#v", loaded)
	}

	event, err := store.LogEvent("drink_water", "2026-05-08", "")
	if err != nil {
		t.Fatalf("LogEvent() for templated habit error = %v", err)
	}
	if event.HabitID != "drink_water" {
		t.Fatalf("LogEvent() habit ID = %q, want drink_water", event.HabitID)
	}
}

func TestStoreDueHabits(t *testing.T) {
	store := NewStore(t.TempDir())
	initTestStore(t, store)
	addTestHabit(t, store, "daily", Schedule{Type: "daily", Target: 1})
	addTestHabit(t, store, "weekly_any", Schedule{Type: "weekly", Target: 3})
	addTestHabit(t, store, "weekly_fri", Schedule{Type: "weekly", Target: 1, Days: []string{"fri"}})
	addTestHabit(t, store, "weekly_mon", Schedule{Type: "weekly", Target: 1, Days: []string{"mon"}})
	addTestHabit(t, store, "monthly", Schedule{Type: "monthly", Target: 1})

	habits, err := store.DueHabits("2026-05-08")
	if err != nil {
		t.Fatalf("DueHabits() error = %v", err)
	}

	got := map[string]bool{}
	for _, habit := range habits {
		got[habit.ID] = true
	}
	for _, id := range []string{"daily", "weekly_any", "weekly_fri", "monthly"} {
		if !got[id] {
			t.Fatalf("DueHabits() missing %q in %#v", id, got)
		}
	}
	if got["weekly_mon"] {
		t.Fatalf("DueHabits() included monday-only habit on friday")
	}
}

func TestStoreReportUsesPeriodTargets(t *testing.T) {
	store := NewStore(t.TempDir())
	initTestStore(t, store)
	addTestHabit(t, store, "exercise", Schedule{Type: "weekly", Target: 3})
	addTestHabit(t, store, "read", Schedule{Type: "daily", Target: 1})

	for _, event := range []Event{
		{ID: "e1", HabitID: "exercise", Date: "2026-05-04", CreatedAt: "2026-05-04T10:00:00Z"},
		{ID: "e2", HabitID: "exercise", Date: "2026-05-06", CreatedAt: "2026-05-06T10:00:00Z"},
		{ID: "r1", HabitID: "read", Date: "2026-05-08", CreatedAt: "2026-05-08T10:00:00Z"},
	} {
		if err := store.AppendEvent(event); err != nil {
			t.Fatalf("AppendEvent(%s) error = %v", event.ID, err)
		}
	}

	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	period, rows, err := store.Report("week", now)
	if err != nil {
		t.Fatalf("Report(week) error = %v", err)
	}
	if period.From != "2026-05-04" || period.To != "2026-05-10" {
		t.Fatalf("Report(week) period = %#v", period)
	}

	byID := map[string]ReportRow{}
	for _, row := range rows {
		byID[row.Habit.ID] = row
	}
	if got := byID["exercise"]; got.Done != 2 || got.Target != 3 {
		t.Fatalf("exercise row = %#v, want done 2 target 3", got)
	}
	if got := byID["read"]; got.Done != 1 || got.Target != 7 {
		t.Fatalf("read row = %#v, want done 1 target 7", got)
	}
}

func TestStoreRejectsInvalidHabit(t *testing.T) {
	store := NewStore(t.TempDir())
	initTestStore(t, store)

	err := store.AddHabit(Habit{
		ID:   "Bad ID",
		Name: "Bad",
		Schedule: Schedule{
			Type:   "daily",
			Target: 1,
		},
	})
	if err == nil {
		t.Fatal("AddHabit() with invalid ID succeeded, want error")
	}

	err = store.AddHabit(Habit{
		ID:   "bad-schedule",
		Name: "Bad schedule",
		Schedule: Schedule{
			Type:   "yearly",
			Target: 1,
		},
	})
	if err == nil {
		t.Fatal("AddHabit() with invalid schedule succeeded, want error")
	}
}

func TestStoreFailsFastWhenRootIsNotInitialized(t *testing.T) {
	store := NewStore(t.TempDir())

	assertNotInitialized(t, store.EnsureInitialized())
	assertNotInitialized(t, store.AddHabit(Habit{
		ID:       "read",
		Name:     "Read",
		Schedule: Schedule{Type: "daily", Target: 1},
	}))
	_, err := store.LoadEvents("", "")
	assertNotInitialized(t, err)
	assertNotInitialized(t, store.AppendEvent(Event{
		ID:        "e1",
		HabitID:   "read",
		Date:      "2026-05-08",
		CreatedAt: "2026-05-08T10:00:00Z",
	}))
	_, _, err = store.Report("week", time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC))
	assertNotInitialized(t, err)
}

func initTestStore(t *testing.T, store Store) {
	t.Helper()
	if err := store.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
}

func assertNotInitialized(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("got nil error, want not initialized error")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("error = %v, want not initialized error", err)
	}
}

func addTestHabit(t *testing.T, store Store, id string, schedule Schedule) {
	t.Helper()
	if err := store.AddHabit(Habit{
		ID:        id,
		Name:      id,
		CreatedAt: "2026-05-01",
		Schedule:  schedule,
	}); err != nil {
		t.Fatalf("AddHabit(%q) error = %v", id, err)
	}
}
