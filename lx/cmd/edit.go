package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/lodek/lx/parse"
	"github.com/lodek/lx/store"
	"github.com/spf13/cobra"
)

func editCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <collection>",
		Short: "Open collection in $EDITOR, then extract done items",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				fatal("EDITOR environment variable is not set")
			}

			collection := args[0]
			path := filepath.Join(s.Dir, collection+".md")
			if _, err := os.Stat(path); err != nil {
				fatal("collection %q not found", collection)
			}

			c := exec.Command(editor, path)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				fatal("editor exited with error: %v", err)
			}

			// After editor returns, extract done items
			content, err := s.ReadFile(collection)
			if err != nil {
				fatal("reading %s: %v", collection, err)
			}

			col := parse.ParseFile(collection, content)
			today := time.Now().Format("2006-01-02")
			var records []store.DoneRecord

			for i := range col.Sublists {
				var kept []parse.Item
				for _, item := range col.Sublists[i].Items {
					if item.Status == "done" {
						records = append(records, store.DoneRecord{
							List:    collection,
							Sublist: col.Sublists[i].Name,
							Date:    today,
							Entry:   item.Text,
						})
					} else {
						kept = append(kept, item)
					}
				}
				col.Sublists[i].Items = kept
			}

			if len(records) == 0 {
				return
			}

			rendered := parse.Render(col)
			if err := s.WriteFile(collection, []byte(rendered)); err != nil {
				fatal("writing %s: %v", collection, err)
			}

			if err := s.AppendDoneLog(records); err != nil {
				fatal("appending to done log: %v", err)
			}

			fmt.Printf("Removed %d completed item(s).\n", len(records))
		},
	}
}
