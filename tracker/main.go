package main

import (
	"fmt"
	"os"

	"tracker/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "tracker:", err)
		os.Exit(1)
	}
}
