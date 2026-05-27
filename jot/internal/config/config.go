package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultConfig = `# jot configuration
# The default collection used when no collection is specified.
# default_collection: personal

# Map collection names to directories containing journal markdown files.
collections:
  # example: ~/journals/personal
`

// Config holds the application configuration.
type Config struct {
	DefaultCollection string            `yaml:"default_collection"`
	Collections       map[string]string `yaml:"collections"` // name -> directory path
}

// ResolveCollection returns the directory for the given collection name.
// If name is empty, it uses the default collection.
func (c *Config) ResolveCollection(name string) (string, string, error) {
	if name == "" {
		if c.DefaultCollection == "" {
			return "", "", fmt.Errorf("no collection specified and no default_collection set in config")
		}
		name = c.DefaultCollection
	}
	dir, ok := c.Collections[name]
	if !ok {
		var names []string
		for n := range c.Collections {
			names = append(names, n)
		}
		return "", "", fmt.Errorf("unknown collection %q (available: %s)", name, joinStrings(names))
	}
	return name, dir, nil
}

func joinStrings(s []string) string {
	result := ""
	for i, v := range s {
		if i > 0 {
			result += ", "
		}
		result += v
	}
	return result
}

// DefaultPath returns the default config file path (~/.config/jot/config.yaml).
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".jot.yaml")
}

// Load reads configuration from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	if cfg.Collections == nil {
		return nil, fmt.Errorf("config must define at least one collection")
	}

	// Expand ~ in paths
	for name, dir := range cfg.Collections {
		if len(dir) > 0 && dir[0] == '~' {
			home, _ := os.UserHomeDir()
			cfg.Collections[name] = filepath.Join(home, dir[1:])
		}
	}

	return &cfg, nil
}

// Init creates a default config file at the given path.
// Returns an error if the file already exists.
func Init(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config already exists at %s", path)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("writing config %s: %w", path, err)
	}

	return nil
}
