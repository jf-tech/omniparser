package cmd

import (
	"strconv"
	"time"

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

type buildInfo struct {
	BuildSHA  string
	BuildTime string
}

var build buildInfo

// Execute executes the root command.
func Execute(commit, epoch string) error {
	epochSec, err := strconv.ParseInt(epoch, 10, 64)
	if err != nil {
		return err
	}
	build = buildInfo{
		BuildSHA:  commit,
		BuildTime: time.Unix(epochSec, 0).UTC().Format(time.RFC3339),
	}
	return rootCmd.Execute()
}
