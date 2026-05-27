package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tracker/internal/tracker"
)

func (a *app) reviewCommand() *cobra.Command {
	var date string
	var editor string

	cmd := &cobra.Command{
		Use:   "review",
		Short: "Review and log all due habits for a day",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := a.store()
			habits, err := store.DueHabits(date)
			if err != nil {
				return err
			}
			if len(habits) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No habits due on %s.\n", date)
				return nil
			}

			checked, err := runReviewEditor(date, habits, editor)
			if err != nil {
				return err
			}
			events, err := store.LogReview(date, checked)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "logged %d habits for %s\n", len(events), date)
			return nil
		},
	}

	cmd.Flags().StringVar(&date, "date", time.Now().AddDate(0, 0, -1).Format(tracker.DateLayout), "review date")
	cmd.Flags().StringVar(&editor, "editor", defaultEditor(), "editor command")
	return cmd
}

func runReviewEditor(date string, habits []tracker.Habit, editor string) ([]string, error) {
	file, err := os.CreateTemp("", "tracker-review-*.md")
	if err != nil {
		return nil, err
	}
	path := file.Name()
	defer os.Remove(path)

	if _, err := file.Write(renderReview(date, habits)); err != nil {
		file.Close()
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}

	parts := strings.Fields(editor)
	if len(parts) == 0 {
		parts = []string{"vim"}
	}
	args := append(parts[1:], path)
	command := exec.Command(parts[0], args...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return nil, err
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseReview(contents), nil
}

func renderReview(date string, habits []tracker.Habit) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# Habit review for %s\n\n", date)
	for _, habit := range habits {
		fmt.Fprintf(&b, "- [ ] %s %s\n", habit.ID, habit.Name)
	}
	return b.Bytes()
}

func parseReview(contents []byte) []string {
	var habitIDs []string
	for _, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)
		if !strings.HasPrefix(lower, "- [x] ") {
			continue
		}
		rest := strings.TrimSpace(line[len("- [x] "):])
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		habitIDs = append(habitIDs, fields[0])
	}
	return habitIDs
}

func defaultEditor() string {
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "vim"
}
