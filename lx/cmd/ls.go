package cmd

import (
	"fmt"

	"github.com/lodek/lx/parse"
	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available collections",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			names, err := s.ListFiles()
			if err != nil {
				fatal("listing collections: %v", err)
			}
			for _, name := range names {
				content, err := s.ReadFile(name)
				if err != nil {
					fatal("reading %s: %v", name, err)
				}
				col := parse.ParseFile(name, content)
				if col.Description != "" {
					fmt.Printf("%s — %s\n", name, col.Description)
				} else {
					fmt.Println(name)
				}
			}
		},
	}
}
