package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "op",
	Long: "op is a CLI of omniparser that transforms data input (such as CSV/XML/JSON/EDI/etc) into desired output by a schema.",
}

func init() {
	rootCmd.AddCommand(transformCmd)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
