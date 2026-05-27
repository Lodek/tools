package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *app) eventsCommand() *cobra.Command {
	var habitID, from, to string

	cmd := &cobra.Command{
		Use:   "events",
		Short: "Query logged events",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			events, err := a.store().LoadEvents(from, to)
			if err != nil {
				return err
			}
			for _, event := range events {
				if habitID != "" && event.HabitID != habitID {
					continue
				}
				if event.Note == "" {
					fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n", event.Date, event.HabitID)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%s  %-20s %s\n", event.Date, event.HabitID, event.Note)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&habitID, "habit", "", "habit id")
	cmd.Flags().StringVar(&from, "from", "", "start date")
	cmd.Flags().StringVar(&to, "to", "", "end date")
	return cmd
}
