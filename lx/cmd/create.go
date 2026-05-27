package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func createCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <collection>",
		Short: "Create a new list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			description, _ := cmd.Flags().GetString("description")

			path := filepath.Join(s.Dir, name+".md")
			if _, err := os.Stat(path); err == nil {
				fatal("collection %q already exists", name)
			}

			var content string
			if description != "" {
				content = fmt.Sprintf("---\ndescription: %s\n---\n", description)
			}

			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				fatal("creating %s: %v", name, err)
			}
		},
	}

	cmd.Flags().StringP("description", "d", "", "description for the collection")
	return cmd
}
