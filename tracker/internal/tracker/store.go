package tracker

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	activeDir   = "habits/active"
	archivedDir = "habits/archived"
	eventsDir   = "events"
)

type Store struct {
	Root string
}

func NewStore(root string) Store {
	if root == "" {
		root = "."
	}
	return Store{Root: root}
}

func (s Store) Init() error {
	for _, dir := range []string{activeDir, archivedDir, eventsDir} {
		if err := os.MkdirAll(s.path(dir), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (s Store) EnsureInitialized() error {
	for _, dir := range []string{activeDir, archivedDir, eventsDir} {
		path := s.path(dir)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return fmt.Errorf("tracker root %q is not initialized: missing %s; run `tracker init`", s.Root, dir)
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("tracker root %q is invalid: %s is not a directory", s.Root, dir)
		}
	}
	return nil
}

func (s Store) AddHabit(habit Habit) error {
	if err := s.EnsureInitialized(); err != nil {
		return err
	}
	if err := ValidateID(habit.ID); err != nil {
		return err
	}
	if strings.TrimSpace(habit.Name) == "" {
		return errors.New("habit name is required")
	}
	if err := validateSchedule(habit.Schedule); err != nil {
		return err
	}
	if habit.CreatedAt == "" {
		habit.CreatedAt = time.Now().Format(DateLayout)
	}
	if habit.Timezone == "" {
		habit.Timezone = time.Local.String()
	}

	activePath := s.habitPath(activeDir, habit.ID)
	if _, err := s.findHabitFile(activeDir, habit.ID); err == nil {
		return fmt.Errorf("active habit %q already exists", habit.ID)
	}
	if _, err := s.findHabitFile(archivedDir, habit.ID); err == nil {
		return fmt.Errorf("archived habit %q already exists", habit.ID)
	}

	return s.writeHabit(activePath, habit)
}

func (s Store) CreateHabitTemplate(id, fileName string) (string, error) {
	if err := s.EnsureInitialized(); err != nil {
		return "", err
	}
	if err := ValidateID(id); err != nil {
		return "", err
	}
	if strings.TrimSpace(fileName) == "" {
		return "", errors.New("file name is required")
	}

	path := s.templatePath(fileName)
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("habit template file %q already exists", path)
	}
	if _, err := s.findHabitFile(activeDir, id); err == nil {
		return "", fmt.Errorf("active habit %q already exists", id)
	}
	if _, err := s.findHabitFile(archivedDir, id); err == nil {
		return "", fmt.Errorf("archived habit %q already exists", id)
	}

	habit := Habit{
		ID:        id,
		Name:      titleFromID(id),
		Timezone:  time.Local.String(),
		CreatedAt: time.Now().Format(DateLayout),
		Schedule: Schedule{
			Type:   "daily",
			Target: 1,
		},
	}
	return path, s.writeHabit(path, habit)
}

func (s Store) ArchiveHabit(id string) error {
	if err := s.EnsureInitialized(); err != nil {
		return err
	}
	src, err := s.findHabitFile(activeDir, id)
	if err != nil {
		return fmt.Errorf("active habit %q not found", id)
	}
	dst := s.path(filepath.Join(archivedDir, filepath.Base(src)))
	return os.Rename(src, dst)
}

func (s Store) LoadHabit(id string) (Habit, error) {
	if err := s.EnsureInitialized(); err != nil {
		return Habit{}, err
	}
	path, err := s.findHabitFile(activeDir, id)
	if err == nil {
		return s.readHabit(path, false)
	}
	if os.IsNotExist(err) {
		if _, archivedErr := s.findHabitFile(archivedDir, id); archivedErr == nil {
			return Habit{}, fmt.Errorf("habit %q is archived", id)
		}
		return Habit{}, fmt.Errorf("active habit %q not found", id)
	}
	return Habit{}, err
}

func (s Store) LoadHabits(includeArchived bool) ([]Habit, error) {
	if err := s.EnsureInitialized(); err != nil {
		return nil, err
	}
	habits, err := s.loadHabitsFromDir(activeDir, false)
	if err != nil {
		return nil, err
	}
	if includeArchived {
		archived, err := s.loadHabitsFromDir(archivedDir, true)
		if err != nil {
			return nil, err
		}
		habits = append(habits, archived...)
	}
	sort.Slice(habits, func(i, j int) bool { return habits[i].ID < habits[j].ID })
	return habits, nil
}

func (s Store) LogEvent(habitID, doneAt, note string) (Event, error) {
	habit, err := s.LoadHabit(habitID)
	if err != nil {
		return Event{}, err
	}
	if _, err := ParseDate(doneAt); err != nil {
		return Event{}, err
	}

	now := time.Now()
	event := Event{
		ID:        NewEventID(habit.ID, now),
		HabitID:   habit.ID,
		Date:      doneAt,
		CreatedAt: now.Format(time.RFC3339),
		Note:      strings.TrimSpace(note),
	}
	if err := s.AppendEvent(event); err != nil {
		return Event{}, err
	}
	return event, nil
}

func (s Store) AppendEvent(event Event) error {
	if err := s.EnsureInitialized(); err != nil {
		return err
	}
	if _, err := ParseDate(event.Date); err != nil {
		return err
	}
	path := s.path(filepath.Join(eventsDir, event.Date[:7]+".jsonl"))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoded, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = file.Write(append(encoded, '\n'))
	return err
}

func (s Store) LogReview(date string, habitIDs []string) ([]Event, error) {
	if _, err := ParseDate(date); err != nil {
		return nil, err
	}
	events := make([]Event, 0, len(habitIDs))
	for _, habitID := range habitIDs {
		event, err := s.LogEvent(habitID, date, "")
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (s Store) LoadEvents(fromValue, toValue string) ([]Event, error) {
	if err := s.EnsureInitialized(); err != nil {
		return nil, err
	}
	from, to, err := parseRange(fromValue, toValue)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.path(eventsDir))
	if err != nil {
		return nil, err
	}

	var events []Event
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		fileEvents, err := s.readEventsFile(s.path(filepath.Join(eventsDir, entry.Name())), from, to)
		if err != nil {
			return nil, err
		}
		events = append(events, fileEvents...)
	}
	SortEvents(events)
	return events, nil
}

func (s Store) DueHabits(date string) ([]Habit, error) {
	day, err := ParseDate(date)
	if err != nil {
		return nil, err
	}
	habits, err := s.LoadHabits(false)
	if err != nil {
		return nil, err
	}
	due := make([]Habit, 0, len(habits))
	for _, habit := range habits {
		if IsDueOn(habit.Schedule, day) {
			due = append(due, habit)
		}
	}
	return due, nil
}

func (s Store) Report(kind string, now time.Time) (Period, []ReportRow, error) {
	from, to := reportBounds(kind, now)
	habits, err := s.LoadHabits(false)
	if err != nil {
		return Period{}, nil, err
	}
	events, err := s.LoadEvents(from.Format(DateLayout), to.Format(DateLayout))
	if err != nil {
		return Period{}, nil, err
	}

	counts := map[string]int{}
	for _, event := range events {
		counts[event.HabitID]++
	}

	rows := make([]ReportRow, 0, len(habits))
	for _, habit := range habits {
		rows = append(rows, ReportRow{
			Habit:  habit,
			Done:   counts[habit.ID],
			Target: TargetForPeriod(habit.Schedule, kind, now),
		})
	}

	return Period{From: from.Format(DateLayout), To: to.Format(DateLayout)}, rows, nil
}

func (s Store) path(parts ...string) string {
	all := append([]string{s.Root}, parts...)
	return filepath.Join(all...)
}

func (s Store) habitPath(dir, id string) string {
	return s.path(filepath.Join(dir, id+".yaml"))
}

func (s Store) templatePath(fileName string) string {
	if filepath.IsAbs(fileName) {
		return fileName
	}
	if !strings.HasSuffix(fileName, ".yaml") && !strings.HasSuffix(fileName, ".yml") {
		fileName += ".yaml"
	}
	if filepath.Dir(fileName) == "." {
		return s.path(filepath.Join(activeDir, fileName))
	}
	return s.path(fileName)
}

func (s Store) findHabitFile(dir, id string) (string, error) {
	path := s.habitPath(dir, id)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	entries, err := os.ReadDir(s.path(dir))
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() || !isHabitFile(entry.Name()) {
			continue
		}
		candidate := s.path(filepath.Join(dir, entry.Name()))
		habit, err := s.readHabit(candidate, false)
		if err != nil {
			return "", err
		}
		if habit.ID == id {
			return candidate, nil
		}
	}
	return "", os.ErrNotExist
}

func (s Store) loadHabitsFromDir(dir string, archived bool) ([]Habit, error) {
	entries, err := os.ReadDir(s.path(dir))
	if err != nil {
		return nil, err
	}

	var habits []Habit
	for _, entry := range entries {
		if entry.IsDir() || !isHabitFile(entry.Name()) {
			continue
		}
		habit, err := s.readHabit(s.path(filepath.Join(dir, entry.Name())), archived)
		if err != nil {
			return nil, err
		}
		habits = append(habits, habit)
	}
	return habits, nil
}

func (s Store) writeHabit(path string, habit Habit) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	encoded, err := yaml.Marshal(habit)
	if err != nil {
		return err
	}
	return os.WriteFile(path, encoded, 0o644)
}

func isHabitFile(name string) bool {
	return strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}

func (s Store) readHabit(path string, archived bool) (Habit, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Habit{}, err
	}

	var habit Habit
	if err := yaml.Unmarshal(contents, &habit); err != nil {
		return Habit{}, err
	}
	if habit.ID == "" || habit.Name == "" || habit.Schedule.Type == "" || habit.Schedule.Target == 0 {
		return Habit{}, fmt.Errorf("%s: missing required habit fields", path)
	}
	if err := validateSchedule(habit.Schedule); err != nil {
		return Habit{}, fmt.Errorf("%s: %w", path, err)
	}
	habit.Archived = archived
	habit.Path = path
	return habit, nil
}

func (s Store) readEventsFile(path string, from, to *time.Time) ([]Event, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []Event
	scanner := bufio.NewScanner(file)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		var event Event
		if err := json.Unmarshal([]byte(text), &event); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, line, err)
		}
		doneAt, err := ParseDate(event.Date)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: invalid date", path, line)
		}
		if from != nil && doneAt.Before(*from) {
			continue
		}
		if to != nil && doneAt.After(*to) {
			continue
		}
		events = append(events, event)
	}
	return events, scanner.Err()
}

func titleFromID(id string) string {
	parts := strings.FieldsFunc(id, func(r rune) bool {
		return r == '-' || r == '_'
	})
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func parseRange(fromValue, toValue string) (*time.Time, *time.Time, error) {
	var from, to *time.Time
	if fromValue != "" {
		parsed, err := ParseDate(fromValue)
		if err != nil {
			return nil, nil, err
		}
		from = &parsed
	}
	if toValue != "" {
		parsed, err := ParseDate(toValue)
		if err != nil {
			return nil, nil, err
		}
		to = &parsed
	}
	return from, to, nil
}

func validateSchedule(schedule Schedule) error {
	if schedule.Target < 1 {
		return errors.New("schedule target must be at least 1")
	}
	switch schedule.Type {
	case "daily", "weekly", "monthly":
		return nil
	default:
		return fmt.Errorf("unsupported schedule type %q", schedule.Type)
	}
}
