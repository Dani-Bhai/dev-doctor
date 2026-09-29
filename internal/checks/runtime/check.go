package runtime

import (
	"fmt"
	"strings"

	"github.com/Dani-Bhai/dev-doctor/internal/model"
)

type ManagerResult struct {
	Manager  string
	Versions []string
}

type (
	NodeManagerResult   = ManagerResult
	JavaManagerResult   = ManagerResult
	PythonManagerResult = ManagerResult
	GoManagerResult     = ManagerResult
)

func runtimeResult(name, tool, activeVersion, activePath string, managers []ManagerResult) model.Result {
	result := model.Result{
		Name:  name,
		Group: "Runtime",
	}

	if activeVersion == "" {
		result.Status = model.StatusError
		result.Message = "not found"

		return result
	}

	result.Status = model.StatusOk
	result.Message = activeVersion
	result.Summary = versionSummary(managers)

	if activePath != "" {
		result.Details = append(result.Details, "  Path: "+activePath)
	}

	if len(managers) > 0 {
		result.Details = append(result.Details, "", "Detected managers:")

		for _, manager := range managers {
			result.Details = append(result.Details, "", manager.Manager)

			for _, version := range manager.Versions {
				result.Details = append(result.Details, "  - "+tool+" "+version)
			}
		}
	}

	return result
}

// versionSummary condenses the manager listings into one line, e.g.
// "  5 versions found (mise, IntelliJ)".
func versionSummary(managers []ManagerResult) []string {
	total := 0
	names := make([]string, 0, len(managers))

	for _, manager := range managers {
		if len(manager.Versions) == 0 {
			continue
		}

		total += len(manager.Versions)
		names = append(names, manager.Manager)
	}

	if total == 0 {
		return nil
	}

	word := "versions"
	if total == 1 {
		word = "version"
	}

	return []string{fmt.Sprintf("  %d %s found (%s)", total, word, strings.Join(names, ", "))}
}
