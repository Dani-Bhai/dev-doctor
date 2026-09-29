package output

import (
	"fmt"

	"github.com/Dani-Bhai/dev-doctor/internal/model"
)

// Print renders the results. Compact mode prints each result's summary,
// verbose mode prints the full details.
func Print(results []model.Result, verbose bool) {
	currentGroup := ""

	for _, result := range results {
		lines := result.Summary
		if verbose {
			lines = result.Details
		}

		if currentGroup != result.Group {
			currentGroup = result.Group
			fmt.Printf("\n%s\n", currentGroup)
		} else if len(lines) > 0 {
			fmt.Println()
		}

		icon := "✓"

		switch result.Status {
		case model.StatusWarning:
			icon = "⚠"

		case model.StatusError:
			icon = "✗"
		}

		fmt.Printf("%s %-20s %s\n",
			icon,
			result.Name,
			result.Message,
		)

		for _, line := range lines {
			fmt.Println(line)
		}
	}
}
