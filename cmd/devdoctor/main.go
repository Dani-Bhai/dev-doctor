package main

import (
	"context"
	"fmt"
	"os"
	goruntime "runtime"

	"github.com/Dani-Bhai/dev-doctor/internal/model"
	"github.com/Dani-Bhai/dev-doctor/internal/output"
	"github.com/Dani-Bhai/dev-doctor/internal/runner"
)

const usage = `devdoctor checks the health of your local dev environment

Usage:
  devdoctor [command]

Without a command devdoctor prints a compact summary of every check.

Commands:
  all      full details for every check
  node     full Node.js details
  java     full Java details
  python   full Python details
  go       full Go details
  git      full Git details
  docker   full Docker details
  disk     full disk usage details
  memory   full memory details
  swap     full swap details
  help     show this help
`

func main() {
	if goruntime.GOOS != "linux" {
		fmt.Fprintf(os.Stderr, "devdoctor currently supports Linux only (running on %s)\n", goruntime.GOOS)
		os.Exit(1)
	}

	args := os.Args[1:]

	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "devdoctor accepts at most one command\n\n%s", usage)
		os.Exit(1)
	}

	ctx := context.Background()

	var (
		results []model.Result
		verbose bool
	)

	switch {
	case len(args) == 0:
		results = runner.Run(ctx, allChecks())

	case args[0] == "help" || args[0] == "-h" || args[0] == "--help":
		fmt.Print(usage)
		return

	case args[0] == "all":
		verbose = true
		results = runner.Run(ctx, allChecks())

	default:
		check, ok := findCheck(args[0])
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", args[0], usage)
			os.Exit(1)
		}

		verbose = true
		results = runner.Run(ctx, []runner.Check{check})
	}

	output.Print(results, verbose)
}
