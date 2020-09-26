package main

import (
	"os"
	"strconv"
	"time"

	"github.com/jf-tech/omniparser/cli/cmd"
)

var (
	// To populate these vars from build/run
	//   go build/run -ldflags "-X main.gitCommit=$(git rev-list -1 HEAD) -X main.buildEpochSec=$(date +%s)" ...
	gitCommit     string
	buildEpochSec string
)

func main() {
	switch gitCommit {
	case "":
		gitCommit = "(unknown commit)"
	default:
		gitCommit = string(([]rune(gitCommit))[:7])
	}
	if buildEpochSec == "" {
		buildEpochSec = strconv.FormatInt(time.Now().Unix(), 10)
	}
	if err := cmd.Execute(gitCommit, buildEpochSec); err != nil {
		os.Exit(1)
	}
}
