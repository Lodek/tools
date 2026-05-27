package parse

import (
	"regexp"
	"strings"
)

type Item struct {
	Raw    string   // original line text
	Text   string   // item text without checkbox markup (tags left in for display)
	Status string   // "todo", "done", "reference"
	Tags   []string // extracted from <tag> patterns
}

type Sublist struct {
	Name   string
	Active bool // true if header was **bold**
	Items  []Item
}

type Collection struct {
	Filename    string    // e.g. "todo" (without .md)
	Description string    // from frontmatter
	Sublists    []Sublist
	Preamble    []string  // non-item, non-header lines before first sublist (frontmatter excluded)
}

var (
	reActiveHeader   = regexp.MustCompile(`^# \*\*(.+)\*\*$`)
	reInactiveHeader = regexp.MustCompile(`^# (.+)$`)
	reDoneItem       = regexp.MustCompile(`^- \[x\] ?(.+)$`)
	reTodoItem       = regexp.MustCompile(`^- \[ ?\] ?(.+)$`)
	reRefItem        = regexp.MustCompile(`^- (.+)$`)
	reTag            = regexp.MustCompile(`<([^>]+)>`)
)

func ParseFile(filename string, content []byte) *Collection {
	body, description := extractFrontmatter(string(content))

	col := &Collection{
		Filename:    filename,
		Description: description,
	}

	var current *Sublist
	lines := strings.Split(body, "\n")

	for _, line := range lines {
		// Active header: # **name**
		if m := reActiveHeader.FindStringSubmatch(line); m != nil {
			if current != nil {
				col.Sublists = append(col.Sublists, *current)
			}
			current = &Sublist{Name: m[1], Active: true}
			continue
		}

		// Inactive header: # name (must check after active to avoid matching bold headers)
		if m := reInactiveHeader.FindStringSubmatch(line); m != nil {
			// Skip if this would also match as active (shouldn't due to order, but be safe)
			if reActiveHeader.MatchString(line) {
				continue
			}
			if current != nil {
				col.Sublists = append(col.Sublists, *current)
			}
			current = &Sublist{Name: m[1], Active: false}
			continue
		}

		// Done item: - [x] text
		if m := reDoneItem.FindStringSubmatch(line); m != nil {
			item := Item{
				Raw:    line,
				Text:   m[1],
				Status: "done",
				Tags:   ExtractTags(m[1]),
			}
			if current != nil {
				current.Items = append(current.Items, item)
			}
			continue
		}

		// Todo item: - [ ] text
		if m := reTodoItem.FindStringSubmatch(line); m != nil {
			item := Item{
				Raw:    line,
				Text:   m[1],
				Status: "todo",
				Tags:   ExtractTags(m[1]),
			}
			if current != nil {
				current.Items = append(current.Items, item)
			}
			continue
		}

		// Reference item: - text
		if m := reRefItem.FindStringSubmatch(line); m != nil {
			item := Item{
				Raw:    line,
				Text:   m[1],
				Status: "reference",
				Tags:   ExtractTags(m[1]),
			}
			if current != nil {
				current.Items = append(current.Items, item)
			}
			continue
		}

		// Unrecognized line
		if current == nil {
			col.Preamble = append(col.Preamble, line)
		}
		// Lines inside a sublist that don't match are silently dropped from items
		// but preserved via Raw line rendering in Render
	}

	if current != nil {
		col.Sublists = append(col.Sublists, *current)
	}

	return col
}

func ExtractTags(text string) []string {
	matches := reTag.FindAllStringSubmatch(text, -1)
	if matches == nil {
		return nil
	}
	tags := make([]string, len(matches))
	for i, m := range matches {
		tags[i] = m[1]
	}
	return tags
}

// Render reconstructs the markdown file from a Collection.
// Unmodified items use their Raw line for byte-identical round-tripping.
func Render(c *Collection) string {
	var b strings.Builder

	// Frontmatter
	if c.Description != "" {
		b.WriteString("---\n")
		b.WriteString("description: " + c.Description + "\n")
		b.WriteString("---\n")
	}

	// Preamble
	for _, line := range c.Preamble {
		b.WriteString(line)
		b.WriteString("\n")
	}

	for _, sub := range c.Sublists {
		if sub.Active {
			b.WriteString("# **" + sub.Name + "**\n")
		} else {
			b.WriteString("# " + sub.Name + "\n")
		}
		for _, item := range sub.Items {
			if item.Raw != "" {
				b.WriteString(item.Raw)
			} else {
				// Newly constructed item (from add command)
				b.WriteString(renderNewItem(item))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderNewItem(item Item) string {
	var line string
	switch item.Status {
	case "todo":
		line = "- [ ] " + item.Text
	case "done":
		line = "- [x] " + item.Text
	default:
		line = "- " + item.Text
	}
	return line
}
