package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tracker/internal/tracker"
)

func (a *app) habitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "habit",
		Short: "Manage habit definitions",
	}

	cmd.AddCommand(
		a.habitAddCommand(),
		a.habitTemplateCommand(),
		a.habitListCommand(),
		a.habitArchiveCommand(),
	)

	return cmd
}

func (a *app) habitAddCommand() *cobra.Command {
	var name, description, days, timezone string
	var daily bool
	var weekly, monthly int

	cmd := &cobra.Command{
		Use:   "add <id>",
		Short: "Add an active habit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedule, err := scheduleFromFlags(daily, weekly, monthly, days)
			if err != nil {
				return err
			}
			return a.store().AddHabit(tracker.Habit{
				ID:          args[0],
				Name:        strings.TrimSpace(name),
				Description: strings.TrimSpace(description),
				Timezone:    timezone,
				Schedule:    schedule,
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "habit name")
	cmd.Flags().StringVar(&description, "description", "", "habit description")
	cmd.Flags().BoolVar(&daily, "daily", false, "daily habit")
	cmd.Flags().IntVar(&weekly, "weekly", 0, "weekly target")
	cmd.Flags().IntVar(&monthly, "monthly", 0, "monthly target")
	cmd.Flags().StringVar(&days, "days", "", "comma-separated weekdays")
	cmd.Flags().StringVar(&timezone, "timezone", time.Local.String(), "IANA timezone")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func (a *app) habitTemplateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "template <id> <file>",
		Short: "Create a new habit YAML template",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := a.store().CreateHabitTemplate(args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", path)
			return nil
		},
	}
}

func (a *app) habitListCommand() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List habits",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			habits, err := a.store().LoadHabits(all)
			if err != nil {
				return err
			}
			if len(habits) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No habits found.")
				return nil
			}
			for _, habit := range habits {
				state := "active"
				if habit.Archived {
					state = "archived"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-10s %s (%s)\n", habit.ID, state, habit.Name, tracker.DescribeSchedule(habit.Schedule))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "include archived habits")
	return cmd
}

func (a *app) habitArchiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <id>",
		Short: "Move an active habit to the archive",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.store().ArchiveHabit(args[0])
		},
	}
}

func scheduleFromFlags(daily bool, weekly, monthly int, days string) (tracker.Schedule, error) {
	selected := 0
	if daily {
		selected++
	}
	if weekly > 0 {
		selected++
	}
	if monthly > 0 {
		selected++
	}
	if selected != 1 {
		return tracker.Schedule{}, fmt.Errorf("choose exactly one of --daily, --weekly n, or --monthly n")
	}
	if daily {
		return tracker.Schedule{Type: "daily", Target: 1}, nil
	}
	if weekly > 0 {
		return tracker.Schedule{Type: "weekly", Target: weekly, Days: tracker.NormalizeDays(days)}, nil
	}
	return tracker.Schedule{Type: "monthly", Target: monthly}, nil
}
