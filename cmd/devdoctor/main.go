package main 


import (
		"context"
		"github.com/Dani-Bhai/dev-doctor/internal/checks/system"
		"github.com/Dani-Bhai/dev-doctor/internal/output"
		"github.com/Dani-Bhai/dev-doctor/internal/runner"
)

func main() {
		ctx := context.Background()

		checks := []runner.Check {
				system.DiskCheck{},
				system.MemCheck{},
				system.SwapCheck{},
		}

		results := runner.Run(ctx, checks)

		output.Print(results)
}
