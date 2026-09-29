package main

import (
	"context"
	"fmt"
	"os"
	goruntime "runtime"

	"github.com/Dani-Bhai/dev-doctor/internal/output"
	"github.com/Dani-Bhai/dev-doctor/internal/runner"
)

func main() {
	if goruntime.GOOS != "linux" {
		fmt.Fprintf(os.Stderr, "devdoctor currently supports Linux only (running on %s)\n", goruntime.GOOS)
		os.Exit(1)
	}

	ctx := context.Background()

	results := runner.Run(ctx, checks())

	output.Print(results)
}
