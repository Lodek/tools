package cmd

import (
	"fmt"
	"strings"

	"github.com/lodek/lx/parse"
	"github.com/spf13/cobra"
)

func getCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [collection[.sublist]]",
		Short: "Read/filter a list",
		Run: func(cmd *cobra.Command, args []string) {
			active, _ := cmd.Flags().GetBool("active")
			tag, _ := cmd.Flags().GetString("tag")
			all, _ := cmd.Flags().GetBool("all")

			if all {
				names, err := s.ListFiles()
				if err != nil {
					fatal("listing collections: %v", err)
				}
				for _, name := range names {
					printCollection(name, "", active, tag)
				}
				return
			}

			if len(args) < 1 {
				fatal("usage: lx get <collection[.sublist]> [--active] [--tag TAG]")
			}

			collection, sublistFilter, _ := strings.Cut(args[0], ".")
			printCollection(collection, sublistFilter, active, tag)
		},
	}

	cmd.Flags().BoolP("active", "a", false, "show only active sublists")
	cmd.Flags().StringP("tag", "t", "", "filter by tag")
	cmd.Flags().Bool("all", false, "show all collections")

	return cmd
}

func printCollection(name, sublistFilter string, active bool, tag string) {
	content, err := s.ReadFile(name)
	if err != nil {
		fatal("reading %s: %v", name, err)
	}
	col := parse.ParseFile(name, content)

	for _, sub := range col.Sublists {
		if sublistFilter != "" && !strings.EqualFold(sub.Name, sublistFilter) {
			continue
		}
		if active && !sub.Active {
			continue
		}

		items := sub.Items
		if tag != "" {
			items = filterByTag(items, tag)
		}

		if len(items) == 0 {
			continue
		}

		if sub.Active {
			fmt.Printf("# **%s**\n", sub.Name)
		} else {
			fmt.Printf("# %s\n", sub.Name)
		}
		for _, item := range items {
			fmt.Println(item.Raw)
		}
		fmt.Println()
	}
}

func filterByTag(items []parse.Item, tag string) []parse.Item {
	var filtered []parse.Item
	for _, item := range items {
		for _, t := range item.Tags {
			if strings.EqualFold(t, tag) {
				filtered = append(filtered, item)
				break
			}
		}
	}
	return filtered
}
