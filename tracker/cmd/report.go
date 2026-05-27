package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func (a *app) reportCommand(kind string) *cobra.Command {
	return &cobra.Command{
		Use:   kind,
		Short: fmt.Sprintf("Show the %s habit report", kind),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			period, rows, err := a.store().Report(kind, time.Now())
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s through %s\n", period.From, period.To)
			for _, row := range rows {
				status := "ok"
				if row.Done < row.Target {
					status = "due"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %2d/%-2d %-4s %s\n", row.Habit.ID, row.Done, row.Target, status, row.Habit.Name)
			}
			return nil
		},
	}
}
