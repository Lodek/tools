package journal

import "time"

// Query defines filters for searching journal entries.
type Query struct {
	DateFrom   *time.Time
	DateTo     *time.Time
	Month      *time.Month
	Year       *int
	ISOWeek    *int
	Tag        *string
	Collection *string
}

// Filter returns entries matching the query.
func Filter(entries []Entry, q Query) []Entry {
	var result []Entry
	for _, e := range entries {
		if !matches(e, q) {
			continue
		}
		result = append(result, e)
	}
	return result
}

func matches(e Entry, q Query) bool {
	if q.DateFrom != nil && e.Date.Before(*q.DateFrom) {
		return false
	}
	if q.DateTo != nil && e.Date.After(*q.DateTo) {
		return false
	}
	if q.Month != nil {
		if q.Year != nil {
			if e.Date.Month() != *q.Month || e.Date.Year() != *q.Year {
				return false
			}
		} else if e.Date.Month() != *q.Month {
			return false
		}
	}
	if q.ISOWeek != nil {
		_, week := e.Date.ISOWeek()
		if week != *q.ISOWeek {
			return false
		}
	}
	if q.Tag != nil && !e.HasTag(*q.Tag) {
		return false
	}
	if q.Collection != nil && e.Collection != *q.Collection {
		return false
	}
	return true
}
