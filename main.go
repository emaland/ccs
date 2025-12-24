package main

import (
	"os"

	"github.com/emaland/ccs/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
