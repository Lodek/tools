package cmd

import (
	"strings"
	"testing"

	"tracker/internal/tracker"
)

func TestRenderAndParseReview(t *testing.T) {
	contents := renderReview("2026-05-08", []tracker.Habit{
		{ID: "exercise", Name: "Exercise"},
		{ID: "read", Name: "Read"},
	})

	text := string(contents)
	if !strings.Contains(text, "# Habit review for 2026-05-08") {
		t.Fatalf("renderReview() missing header: %s", text)
	}
	if !strings.Contains(text, "- [ ] exercise Exercise") {
		t.Fatalf("renderReview() missing exercise item: %s", text)
	}

	edited := strings.Replace(text, "- [ ] exercise", "- [x] exercise", 1)
	got := parseReview([]byte(edited))
	if len(got) != 1 || got[0] != "exercise" {
		t.Fatalf("parseReview() = %#v, want [exercise]", got)
	}
}
