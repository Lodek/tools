package journal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// Entry represents a single journal entry.
type Entry struct {
	Date       time.Time
	Tags       []string
	Body       string
	Collection string
}

var (
	// Matches lines like: 2024-03-11 [tag1] [tag2]
	headerDateRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})\s*(.*)$`)
	// Matches setext underline (=== or ---)
	setextRe = regexp.MustCompile(`^[=-]{3,}\s*$`)
	// Matches tags like [tag1]
	tagRe = regexp.MustCompile(`\[([^\]]+)\]`)
)

// ISOWeek returns the ISO week number for the entry.
func (e Entry) ISOWeek() int {
	_, week := e.Date.ISOWeek()
	return week
}

// HasTag checks whether the entry has a given tag (case-insensitive).
func (e Entry) HasTag(tag string) bool {
	tag = strings.ToLower(tag)
	for _, t := range e.Tags {
		if strings.ToLower(t) == tag {
			return true
		}
	}
	return false
}

// ParseFile reads a markdown journal file and returns all entries found.
func ParseFile(path, collection string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	return ParseLines(lines, collection)
}

// ParseLines parses journal entries from a slice of lines.
func ParseLines(lines []string, collection string) ([]Entry, error) {
	var entries []Entry
	i := 0

	for i < len(lines) {
		// Look for a date header line followed by a setext underline
		if i+1 < len(lines) && headerDateRe.MatchString(lines[i]) && setextRe.MatchString(lines[i+1]) {
			date, tags, err := parseHeader(lines[i])
			if err != nil {
				i++
				continue
			}

			// Skip past the underline
			i += 2

			// Collect body lines until next header or EOF
			var bodyLines []string
			for i < len(lines) {
				// Peek ahead: is this line a date header with a setext underline following?
				if i+1 < len(lines) && headerDateRe.MatchString(lines[i]) && setextRe.MatchString(lines[i+1]) {
					break
				}
				bodyLines = append(bodyLines, lines[i])
				i++
			}

			body := strings.TrimSpace(strings.Join(bodyLines, "\n"))
			entries = append(entries, Entry{
				Date:       date,
				Tags:       tags,
				Body:       body,
				Collection: collection,
			})
		} else {
			i++
		}
	}

	return entries, nil
}

func parseHeader(line string) (time.Time, []string, error) {
	m := headerDateRe.FindStringSubmatch(line)
	if m == nil {
		return time.Time{}, nil, fmt.Errorf("not a header line: %s", line)
	}

	date, err := time.Parse("2006-01-02", m[1])
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("parsing date %q: %w", m[1], err)
	}

	var tags []string
	if m[2] != "" {
		tagMatches := tagRe.FindAllStringSubmatch(m[2], -1)
		for _, tm := range tagMatches {
			tags = append(tags, tm[1])
		}
	}

	return date, tags, nil
}

// FormatEntry formats an entry as markdown text suitable for writing to a file.
func FormatEntry(e Entry) string {
	header := e.Date.Format("2006-01-02")
	for _, tag := range e.Tags {
		header += " [" + tag + "]"
	}
	underline := strings.Repeat("=", len(header))
	return header + "\n" + underline + "\n\n" + e.Body + "\n"
}

// MonthFile returns the expected filename for a given date (e.g., "2024-03.md").
func MonthFile(date time.Time) string {
	return date.Format("2006-01") + ".md"
}
