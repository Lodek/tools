package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Store struct {
	Dir string
}

type DoneRecord struct {
	List    string `json:"list"`
	Sublist string `json:"sublist,omitempty"`
	Date    string `json:"date"`
	Entry   string `json:"entry"`
}

// ListFiles returns the names of all .md files in the store directory,
// without the .md extension.
func (s *Store) ListFiles() ([]string, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".md") {
			names = append(names, strings.TrimSuffix(e.Name(), ".md"))
		}
	}
	return names, nil
}

// ReadFile reads the content of a collection file.
func (s *Store) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.Dir, name+".md"))
}

// WriteFile writes content to a collection file.
func (s *Store) WriteFile(name string, content []byte) error {
	return os.WriteFile(filepath.Join(s.Dir, name+".md"), content, 0644)
}

// AppendDoneLog appends done records as JSONL to lx.json in the store directory.
func (s *Store) AppendDoneLog(records []DoneRecord) error {
	f, err := os.OpenFile(filepath.Join(s.Dir, "lx.json"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, r := range records {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}
