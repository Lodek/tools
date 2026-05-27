package cmd

import (
	"fmt"
	"os"

	"github.com/lodek/lx/store"
	"github.com/spf13/cobra"
)

var s *store.Store

var rootCmd = &cobra.Command{
	Use:   "lx",
	Short: "CLI for managing flat markdown list files",
}

func Execute() {
	dir := os.Getenv("LX_DIR")
	if dir == "" {
		fatal("LX_DIR environment variable is not set")
	}
	s = &store.Store{Dir: dir}

	rootCmd.AddCommand(listCmd(), getCmd(), editCmd(), addCmd(), doneCmd(), createCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "lx: "+format+"\n", args...)
	os.Exit(1)
}
