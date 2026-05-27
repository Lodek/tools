package tracker

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func DescribeSchedule(schedule Schedule) string {
	switch schedule.Type {
	case "daily":
		return "daily"
	case "weekly":
		if len(schedule.Days) > 0 {
			return fmt.Sprintf("%d/week on %s", schedule.Target, strings.Join(schedule.Days, ","))
		}
		return fmt.Sprintf("%d/week", schedule.Target)
	case "monthly":
		return fmt.Sprintf("%d/month", schedule.Target)
	default:
		return "unknown"
	}
}

func SortEvents(events []Event) {
	sort.Slice(events, func(i, j int) bool {
		if events[i].Date == events[j].Date {
			return events[i].CreatedAt < events[j].CreatedAt
		}
		return events[i].Date < events[j].Date
	})
}

func IsDueOn(schedule Schedule, day time.Time) bool {
	switch schedule.Type {
	case "daily":
		return true
	case "weekly":
		if len(schedule.Days) == 0 {
			return true
		}
		weekday := strings.ToLower(day.Weekday().String()[:3])
		for _, dueDay := range schedule.Days {
			if strings.ToLower(dueDay) == weekday {
				return true
			}
		}
		return false
	case "monthly":
		return true
	default:
		return false
	}
}

func TargetForPeriod(schedule Schedule, kind string, now time.Time) int {
	switch kind {
	case "today":
		if schedule.Type == "daily" {
			return schedule.Target
		}
	case "week":
		if schedule.Type == "daily" {
			return 7 * schedule.Target
		}
		if schedule.Type == "weekly" {
			return schedule.Target
		}
	case "month":
		if schedule.Type == "daily" {
			return DaysInMonth(now) * schedule.Target
		}
		if schedule.Type == "weekly" {
			return schedule.Target * WeeksTouchedByMonth(now)
		}
		if schedule.Type == "monthly" {
			return schedule.Target
		}
	}
	return 0
}

func reportBounds(kind string, now time.Time) (time.Time, time.Time) {
	switch kind {
	case "today":
		from := DateOnly(now)
		return from, from
	case "week":
		from := StartOfWeek(now)
		return from, from.AddDate(0, 0, 6)
	case "month":
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return from, from.AddDate(0, 1, -1)
	default:
		from := DateOnly(now)
		return from, from
	}
}
