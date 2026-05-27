package tracker

import "encoding/json"

type Habit struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Timezone    string   `yaml:"timezone,omitempty"`
	CreatedAt   string   `yaml:"created_at"`
	Schedule    Schedule `yaml:"schedule"`

	Archived bool   `yaml:"-"`
	Path     string `yaml:"-"`
}

type Schedule struct {
	Type   string   `yaml:"type"`
	Target int      `yaml:"target"`
	Days   []string `yaml:"days,omitempty"`
}

type Event struct {
	ID        string `json:"id"`
	HabitID   string `json:"habit_id"`
	Date      string `json:"date"`
	CreatedAt string `json:"created_at"`
	Note      string `json:"note,omitempty"`
}

func (e *Event) UnmarshalJSON(data []byte) error {
	type event Event
	var raw struct {
		event
		DoneAt string `json:"done_at"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*e = Event(raw.event)
	if e.Date == "" {
		e.Date = raw.DoneAt
	}
	return nil
}

type Period struct {
	From string
	To   string
}

type ReportRow struct {
	Habit  Habit
	Done   int
	Target int
}
