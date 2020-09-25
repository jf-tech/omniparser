package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "op",
	Long: "op is a CLI of omniparser that ingests data input (such as CSV/XML/JSON/EDI/etc) and transforms into desired output by a schema.",
}

func init() {
	rootCmd.AddCommand(transformCmd)
	rootCmd.AddCommand(serverCmd)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
