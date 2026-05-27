package tracker

import (
	"fmt"
	"time"
)

const DateLayout = "2006-01-02"

func ParseDate(value string) (time.Time, error) {
	parsed, err := time.ParseInLocation(DateLayout, value, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("date must use YYYY-MM-DD: %q", value)
	}
	return parsed, nil
}

func DateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func StartOfWeek(t time.Time) time.Time {
	d := DateOnly(t)
	offset := (int(d.Weekday()) + 6) % 7
	return d.AddDate(0, 0, -offset)
}

func DaysInMonth(t time.Time) int {
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	return start.AddDate(0, 1, -1).Day()
}

func WeeksTouchedByMonth(t time.Time) int {
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	end := start.AddDate(0, 1, -1)
	return int(end.Sub(start).Hours()/24)/7 + 1
}
