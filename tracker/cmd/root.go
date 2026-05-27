package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"tracker/internal/tracker"
)

type app struct {
	root string
}

func Execute() error {
	a := &app{root: defaultRoot()}
	return a.rootCommand().Execute()
}

func (a *app) rootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "tracker",
		Short:         "A lightweight file-backed habit tracker",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.PersistentFlags().StringVar(&a.root, "root", a.root, "tracker root directory")

	root.AddCommand(
		a.initCommand(),
		a.habitCommand(),
		a.logCommand(),
		a.reviewCommand(),
		a.eventsCommand(),
		a.reportCommand("today"),
		a.reportCommand("week"),
		a.reportCommand("month"),
	)

	return root
}

func (a *app) store() tracker.Store {
	return tracker.NewStore(a.root)
}

func defaultRoot() string {
	if root := os.Getenv("TRACKER_ROOT"); root != "" {
		return root
	}
	return "."
}

func (a *app) initCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create tracker directories",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.store().Init()
		},
	}
}
