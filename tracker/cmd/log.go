package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"tracker/internal/tracker"
)

func (a *app) logCommand() *cobra.Command {
	var date, note string

	cmd := &cobra.Command{
		Use:   "log <habit-id>",
		Short: "Log a habit completion",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			event, err := a.store().LogEvent(args[0], date, note)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "logged %s on %s\n", event.HabitID, event.Date)
			return nil
		},
	}

	cmd.Flags().StringVar(&date, "date", time.Now().Format(tracker.DateLayout), "completion date")
	cmd.Flags().StringVar(&note, "note", "", "optional note")
	return cmd
}
