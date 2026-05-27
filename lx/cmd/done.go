package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lodek/lx/parse"
	"github.com/lodek/lx/store"
	"github.com/spf13/cobra"
)

func doneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done",
		Short: "Clean up completed items across all collections",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			names, err := s.ListFiles()
			if err != nil {
				fatal("listing collections: %v", err)
			}

			today := time.Now().Format("2006-01-02")
			var allRecords []store.DoneRecord

			for _, name := range names {
				content, err := s.ReadFile(name)
				if err != nil {
					fatal("reading %s: %v", name, err)
				}

				col := parse.ParseFile(name, content)
				var records []store.DoneRecord

				for i := range col.Sublists {
					var kept []parse.Item
					for _, item := range col.Sublists[i].Items {
						if item.Status == "done" {
							records = append(records, store.DoneRecord{
								List:    name,
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
					continue
				}

				allRecords = append(allRecords, records...)

				if !dryRun {
					rendered := parse.Render(col)
					if err := s.WriteFile(name, []byte(rendered)); err != nil {
						fatal("writing %s: %v", name, err)
					}
				}
			}

			if len(allRecords) == 0 {
				fmt.Println("No completed items found.")
				return
			}

			if dryRun {
				enc := json.NewEncoder(os.Stdout)
				for _, r := range allRecords {
					enc.Encode(r)
				}
				return
			}

			if err := s.AppendDoneLog(allRecords); err != nil {
				fatal("appending to done log: %v", err)
			}

			fmt.Printf("Removed %d completed item(s).\n", len(allRecords))
		},
	}

	cmd.Flags().Bool("dry-run", false, "show what would be removed without changing the file")
	return cmd
}
