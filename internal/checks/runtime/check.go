package runtime

import "github.com/Dani-Bhai/dev-doctor/internal/model"

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
