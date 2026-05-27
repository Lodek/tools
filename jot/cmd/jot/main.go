package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lodek/jot/internal/config"
	"github.com/lodek/jot/internal/journal"
	"github.com/spf13/cobra"
)

var cfgPath string

func main() {
	root := &cobra.Command{
		Use:   "jot",
		Short: "Journal management tool",
	}

	root.PersistentFlags().StringVar(&cfgPath, "config", config.DefaultPath(), "config file path")

	root.AddCommand(initCmd())
	root.AddCommand(readCmd())
	root.AddCommand(newCmd())
	root.AddCommand(addCmd())
	root.AddCommand(ingestCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a default config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Init(cfgPath); err != nil {
				return err
			}
			fmt.Printf("Config created at %s\n", cfgPath)
			fmt.Println("Edit it to add your journal collections.")
			return nil
		},
	}
}

func loadConfig() *config.Config {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

func readCmd() *cobra.Command {
	var (
		from       string
		to         string
		month      string
		week       int
		tag        string
		collection string
	)

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read and query journal entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfig()

			entries, err := journal.LoadAll(cfg.Collections)
			if err != nil {
				return err
			}

			q := journal.Query{}

			if from != "" {
				t, err := time.Parse("2006-01-02", from)
				if err != nil {
					return fmt.Errorf("invalid --from date: %w", err)
				}
				q.DateFrom = &t
			}
			if to != "" {
				t, err := time.Parse("2006-01-02", to)
				if err != nil {
					return fmt.Errorf("invalid --to date: %w", err)
				}
				q.DateTo = &t
			}
			if month != "" {
				parts := strings.Split(month, "-")
				if len(parts) == 2 {
					y, err := strconv.Atoi(parts[0])
					if err != nil {
						return fmt.Errorf("invalid --month year: %w", err)
					}
					m, err := strconv.Atoi(parts[1])
					if err != nil || m < 1 || m > 12 {
						return fmt.Errorf("invalid --month month: must be 1-12")
					}
					mon := time.Month(m)
					q.Month = &mon
					q.Year = &y
				} else {
					return fmt.Errorf("--month must be YYYY-MM format")
				}
			}
			if week > 0 {
				q.ISOWeek = &week
			}
			if tag != "" {
				q.Tag = &tag
			}
			if collection != "" {
				q.Collection = &collection
			}

			results := journal.Filter(entries, q)

			for i, e := range results {
				if i > 0 {
					fmt.Println()
				}
				header := e.Date.Format("2006-01-02")
				if collection == "" {
					header += fmt.Sprintf(" (%s)", e.Collection)
				}
				for _, t := range e.Tags {
					header += fmt.Sprintf(" [%s]", t)
				}
				fmt.Println(header)
				fmt.Println(strings.Repeat("=", len(header)))
				fmt.Println()
				fmt.Println(e.Body)
			}

			if len(results) == 0 {
				fmt.Println("No entries found.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "end date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&month, "month", "", "filter by month (YYYY-MM)")
	cmd.Flags().IntVar(&week, "week", 0, "filter by ISO week number")
	cmd.Flags().StringVar(&tag, "tag", "", "filter by tag")
	cmd.Flags().StringVar(&collection, "collection", "", "filter by collection name")

	return cmd
}

func newCmd() *cobra.Command {
	var tags []string

	cmd := &cobra.Command{
		Use:   "new [collection]",
		Short: "New creates a new journal entry",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfig()
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			collName, dir, err := cfg.ResolveCollection(name)
			if err != nil {
				return err
			}

			return journal.NewEntry(dir, collName, tags)
		},
	}

	cmd.Flags().StringSliceVar(&tags, "tag", nil, "tags for the entry (can be repeated)")

	return cmd
}

func addCmd() *cobra.Command {
	var (
		tags    []string
		dateStr string
	)

	cmd := &cobra.Command{
		Use:   "add [collection]",
		Short: "Add a journal entry from stdin",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfig()
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			collName, dir, err := cfg.ResolveCollection(name)
			if err != nil {
				return err
			}

			date := time.Now()
			if dateStr != "" {
				date, err = time.Parse("2006-01-02", dateStr)
				if err != nil {
					return fmt.Errorf("invalid --date: %w", err)
				}
			}

			body, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}

			if err := journal.AddEntry(dir, collName, date, tags, string(body)); err != nil {
				return err
			}

			fmt.Printf("Added entry for %s to %q\n", date.Format("2006-01-02"), collName)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&tags, "tag", nil, "tags for the entry (can be repeated)")
	cmd.Flags().StringVar(&dateStr, "date", "", "entry date (YYYY-MM-DD, defaults to today)")

	return cmd
}

func ingestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingest [collection] [file...]",
		Short: "Ingest markdown files into a collection",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfig()

			// Try to resolve args[0] as a collection name. If it fails
			// and there's a default, treat all args as file paths.
			var collName, dir string
			var files []string
			name, d, err := cfg.ResolveCollection(args[0])
			if err == nil {
				collName, dir = name, d
				files = args[1:]
			} else {
				// args[0] isn't a collection — use default
				collName, dir, err = cfg.ResolveCollection("")
				if err != nil {
					return err
				}
				files = args
			}

			if len(files) == 0 {
				return fmt.Errorf("no files specified")
			}

			for _, path := range files {
				n, err := journal.IngestFile(path, dir, collName)
				if err != nil {
					return fmt.Errorf("ingesting %s: %w", path, err)
				}
				fmt.Printf("Ingested %d entries from %s into %q\n", n, path, collName)
			}

			return nil
		},
	}

	return cmd
}
