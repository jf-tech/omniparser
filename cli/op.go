package main

import (
	"os"

	"github.com/jf-tech/omniparser/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
