package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LoadCollection loads all entries from a single collection directory.
func LoadCollection(name, dir string) ([]Entry, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading collection %q dir %s: %w", name, dir, err)
	}

	var entries []Entry
	for _, de := range dirEntries {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, de.Name())
		fileEntries, err := ParseFile(path, name)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		entries = append(entries, fileEntries...)
	}

	return entries, nil
}

// LoadAll loads entries from all collections.
func LoadAll(collections map[string]string) ([]Entry, error) {
	var all []Entry
	for name, dir := range collections {
		entries, err := LoadCollection(name, dir)
		if err != nil {
			return nil, err
		}
		all = append(all, entries...)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Date.Before(all[j].Date)
	})

	return all, nil
}
