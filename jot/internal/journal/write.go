package journal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// NewEntry opens $EDITOR for the user to write a journal entry, then appends
// it to the correct month file in the collection directory.
func NewEntry(collectionDir string, collection string, tags []string) error {
	now := time.Now()
	date := now.Format("2006-01-02")

	// Build header template
	header := date
	for _, tag := range tags {
		header += " [" + tag + "]"
	}
	underline := strings.Repeat("=", len(header))
	template := header + "\n" + underline + "\n\n"

	// Write template to temp file
	tmpFile, err := os.CreateTemp("", "jot-*.md")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(template); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing template: %w", err)
	}
	tmpFile.Close()

	// Open editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	// Read back the edited content
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("reading edited file: %w", err)
	}

	edited := strings.TrimSpace(string(content))
	if edited == "" || edited == strings.TrimSpace(template) {
		fmt.Println("Empty entry, aborting.")
		return nil
	}

	// Parse to validate
	lines := strings.Split(string(content), "\n")
	entries, err := ParseLines(lines, collection)
	if err != nil || len(entries) == 0 {
		return fmt.Errorf("could not parse entry from editor content")
	}

	// Append to the correct month file
	return appendToMonthFile(collectionDir, entries[0])
}

// AddEntry creates a journal entry from the given body text and appends it
// to the correct month file. Unlike NewEntry, this does not open an editor.
func AddEntry(collectionDir, collection string, date time.Time, tags []string, body string) error {
	body = strings.TrimSpace(body)
	if body == "" {
		return fmt.Errorf("empty entry body")
	}

	e := Entry{
		Date:       date,
		Tags:       tags,
		Body:       body,
		Collection: collection,
	}

	return appendToMonthFile(collectionDir, e)
}

// IngestFile parses a markdown file and merges its entries into the collection.
func IngestFile(path, collectionDir, collection string) (int, error) {
	entries, err := ParseFile(path, collection)
	if err != nil {
		return 0, err
	}

	for _, e := range entries {
		if err := appendToMonthFile(collectionDir, e); err != nil {
			return 0, fmt.Errorf("writing entry %s: %w", e.Date.Format("2006-01-02"), err)
		}
	}

	return len(entries), nil
}

func appendToMonthFile(dir string, e Entry) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	filename := MonthFile(e.Date)
	path := filepath.Join(dir, filename)

	// Check if file exists and has content — add separator newlines
	var prefix string
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		// Read existing to check if it ends with newline
		existing, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n\n") {
			if strings.HasSuffix(string(existing), "\n") {
				prefix = "\n"
			} else {
				prefix = "\n\n"
			}
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	content := prefix + FormatEntry(e)
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("writing to %s: %w", path, err)
	}

	return nil
}
