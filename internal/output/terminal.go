package output

import (
		"fmt"
		"github.com/Dani-Bhai/dev-doctor/internal/model"

)

func Print(results []model.Result){
		currentGroup := ""

		for _, result := range results {
				if currentGroup != result.Group {
						currentGroup = result.Group 
						fmt.Printf("\n%s\n", currentGroup)
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
		}
}
