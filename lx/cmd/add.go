package cmd

import (
	"strings"

	"github.com/lodek/lx/parse"
	"github.com/spf13/cobra"
)

func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <collection> <sublist> <item>",
		Short: "Add an item to a sublist",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			collection := args[0]
			sublistName := args[1]
			itemText := args[2]

			tag, _ := cmd.Flags().GetString("tag")
			if tag != "" {
				itemText += " <" + tag + ">"
			}

			content, err := s.ReadFile(collection)
			if err != nil {
				fatal("reading %s: %v", collection, err)
			}

			col := parse.ParseFile(collection, content)

			found := false
			for i := range col.Sublists {
				if strings.EqualFold(col.Sublists[i].Name, sublistName) {
					col.Sublists[i].Items = append(col.Sublists[i].Items, parse.Item{
						Text:   itemText,
						Status: "todo",
						Tags:   parse.ExtractTags(itemText),
					})
					found = true
					break
				}
			}

			if !found {
				col.Sublists = append(col.Sublists, parse.Sublist{
					Name:   sublistName,
					Active: false,
					Items: []parse.Item{
						{
							Text:   itemText,
							Status: "todo",
							Tags:   parse.ExtractTags(itemText),
						},
					},
				})
			}

			rendered := parse.Render(col)
			if err := s.WriteFile(collection, []byte(rendered)); err != nil {
				fatal("writing %s: %v", collection, err)
			}
		},
	}

	cmd.Flags().StringP("tag", "t", "", "tag to add to the item")
	return cmd
}
