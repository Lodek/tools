package journal

import (
	"testing"
	"time"
)

func TestParseFile(t *testing.T) {
	entries, err := ParseFile("../../2024-03.md", "test")
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	expected := time.Date(2024, 3, 11, 0, 0, 0, 0, time.UTC)
	for i, e := range entries {
		if !e.Date.Equal(expected) {
			t.Errorf("entry %d: date = %v, want %v", i, e.Date, expected)
		}
		if e.Collection != "test" {
			t.Errorf("entry %d: collection = %q, want %q", i, e.Collection, "test")
		}
		if e.Body == "" {
			t.Errorf("entry %d: empty body", i)
		}
	}
}

func TestParseHeaderWithTags(t *testing.T) {
	lines := []string{
		"2024-05-10 [reflection] [climbing]",
		"==================================",
		"",
		"Today was a good day.",
	}

	entries, err := ParseLines(lines, "test")
	if err != nil {
		t.Fatalf("ParseLines: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	e := entries[0]
	if len(e.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d: %v", len(e.Tags), e.Tags)
	}
	if e.Tags[0] != "reflection" || e.Tags[1] != "climbing" {
		t.Errorf("tags = %v, want [reflection, climbing]", e.Tags)
	}
}

func TestISOWeek(t *testing.T) {
	e := Entry{Date: time.Date(2024, 3, 11, 0, 0, 0, 0, time.UTC)}
	// 2024-03-11 is ISO week 11
	if w := e.ISOWeek(); w != 11 {
		t.Errorf("ISOWeek = %d, want 11", w)
	}
}

func TestFilter(t *testing.T) {
	entries := []Entry{
		{Date: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), Tags: []string{"work"}, Collection: "personal"},
		{Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), Tags: []string{"climbing"}, Collection: "personal"},
		{Date: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC), Collection: "work"},
	}

	tag := "climbing"
	results := Filter(entries, Query{Tag: &tag})
	if len(results) != 1 {
		t.Errorf("tag filter: got %d, want 1", len(results))
	}

	march := time.March
	year := 2024
	results = Filter(entries, Query{Month: &march, Year: &year})
	if len(results) != 2 {
		t.Errorf("month filter: got %d, want 2", len(results))
	}

	coll := "work"
	results = Filter(entries, Query{Collection: &coll})
	if len(results) != 1 {
		t.Errorf("collection filter: got %d, want 1", len(results))
	}
}

func TestFormatEntry(t *testing.T) {
	e := Entry{
		Date: time.Date(2024, 5, 10, 0, 0, 0, 0, time.UTC),
		Tags: []string{"reflection"},
		Body: "Today was good.",
	}

	got := FormatEntry(e)
	expected := "2024-05-10 [reflection]\n=======================\n\nToday was good.\n"
	if got != expected {
		t.Errorf("FormatEntry:\ngot:  %q\nwant: %q", got, expected)
	}
}
