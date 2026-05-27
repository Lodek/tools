package parse

import (
	"strings"
	"testing"
)

const testFile = `---
description: Personal todo list
---

# **eurotrip**
- [ ] Book flights <this-week>
- [ ] Spain car rental <this-month>
- [x] Pack bags

# **house**
- [ ] Finish pergola <this-week>
- [ ] Climbing wall build

# parked
- [ ] OpenClaw agent layer
- [ ] Antithesis interview
`

const refFile = `---
description: Leisure stuff
---

# **restaurants**
- Korean place on Rua XV
- Sushi spot downtown

# movies
- Blade Runner 2049
- Arrival
`

func TestParseFile_BasicStructure(t *testing.T) {
	col := ParseFile("todo", []byte(testFile))

	if col.Filename != "todo" {
		t.Errorf("Filename = %q, want %q", col.Filename, "todo")
	}
	if col.Description != "Personal todo list" {
		t.Errorf("Description = %q, want %q", col.Description, "Personal todo list")
	}
	if len(col.Sublists) != 3 {
		t.Fatalf("got %d sublists, want 3", len(col.Sublists))
	}
}

func TestParseFile_ActiveInactive(t *testing.T) {
	col := ParseFile("todo", []byte(testFile))

	if !col.Sublists[0].Active {
		t.Error("eurotrip should be active")
	}
	if col.Sublists[0].Name != "eurotrip" {
		t.Errorf("first sublist name = %q, want %q", col.Sublists[0].Name, "eurotrip")
	}
	if !col.Sublists[1].Active {
		t.Error("house should be active")
	}
	if col.Sublists[2].Active {
		t.Error("parked should be inactive")
	}
	if col.Sublists[2].Name != "parked" {
		t.Errorf("third sublist name = %q, want %q", col.Sublists[2].Name, "parked")
	}
}

func TestParseFile_Items(t *testing.T) {
	col := ParseFile("todo", []byte(testFile))

	eurotrip := col.Sublists[0]
	if len(eurotrip.Items) != 3 {
		t.Fatalf("eurotrip has %d items, want 3", len(eurotrip.Items))
	}

	// Todo item with tag
	item := eurotrip.Items[0]
	if item.Status != "todo" {
		t.Errorf("item 0 status = %q, want %q", item.Status, "todo")
	}
	if len(item.Tags) != 1 || item.Tags[0] != "this-week" {
		t.Errorf("item 0 tags = %v, want [this-week]", item.Tags)
	}

	// Done item
	item = eurotrip.Items[2]
	if item.Status != "done" {
		t.Errorf("item 2 status = %q, want %q", item.Status, "done")
	}
	if item.Text != "Pack bags" {
		t.Errorf("item 2 text = %q, want %q", item.Text, "Pack bags")
	}
}

func TestParseFile_ReferenceItems(t *testing.T) {
	col := ParseFile("leisure", []byte(refFile))

	if len(col.Sublists) != 2 {
		t.Fatalf("got %d sublists, want 2", len(col.Sublists))
	}

	restaurants := col.Sublists[0]
	if len(restaurants.Items) != 2 {
		t.Fatalf("restaurants has %d items, want 2", len(restaurants.Items))
	}
	if restaurants.Items[0].Status != "reference" {
		t.Errorf("item status = %q, want %q", restaurants.Items[0].Status, "reference")
	}
	if restaurants.Items[0].Text != "Korean place on Rua XV" {
		t.Errorf("item text = %q", restaurants.Items[0].Text)
	}
}

func TestParseFile_NoFrontmatter(t *testing.T) {
	input := `# **stuff**
- [ ] Do thing
`
	col := ParseFile("bare", []byte(input))
	if col.Description != "" {
		t.Errorf("Description = %q, want empty", col.Description)
	}
	if len(col.Sublists) != 1 {
		t.Fatalf("got %d sublists, want 1", len(col.Sublists))
	}
}

func TestRender_RoundTrip(t *testing.T) {
	col := ParseFile("todo", []byte(testFile))
	rendered := Render(col)

	// Normalize: trim trailing whitespace from both
	want := strings.TrimRight(testFile, "\n") + "\n"
	got := strings.TrimRight(rendered, "\n") + "\n"

	if got != want {
		t.Errorf("round-trip mismatch.\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestRender_RoundTrip_Reference(t *testing.T) {
	col := ParseFile("leisure", []byte(refFile))
	rendered := Render(col)

	want := strings.TrimRight(refFile, "\n") + "\n"
	got := strings.TrimRight(rendered, "\n") + "\n"

	if got != want {
		t.Errorf("round-trip mismatch.\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestExtractTags(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"Book flights <this-week>", []string{"this-week"}},
		{"No tags here", nil},
		{"Multiple <a> tags <b>", []string{"a", "b"}},
		{"<before-eurotrip> plan", []string{"before-eurotrip"}},
	}

	for _, tt := range tests {
		got := ExtractTags(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("ExtractTags(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("ExtractTags(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestExtractFrontmatter(t *testing.T) {
	body, desc := extractFrontmatter("---\ndescription: My list\n---\n# **stuff**\n- item")
	if desc != "My list" {
		t.Errorf("description = %q, want %q", desc, "My list")
	}
	if !strings.Contains(body, "# **stuff**") {
		t.Errorf("body should contain header, got: %q", body)
	}
}
