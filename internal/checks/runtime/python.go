package runtime

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Dani-Bhai/dev-doctor/internal/checks/system"
	"github.com/Dani-Bhai/dev-doctor/internal/model"
)

type PythonInfo struct {
	ActiveVersion string
	ActivePath    string
	Managers      []PythonManagerResult
}

var pythonVersionPattern = regexp.MustCompile(`Python\s+(\S+)`)

func detectPythonManagers() []string {
	var managers []string

	if system.CommandExists("mise") {
		managers = append(managers, "mise")
	}

	if system.CommandExists("pyenv") {
		managers = append(managers, "pyenv")
	}

	if system.CommandExists("asdf") {
		managers = append(managers, "asdf")
	}

	if system.CommandExists("uv") {
		managers = append(managers, "uv")
	}

	return managers
}

func parsePythonVersion(output string) (string, error) {
	matches := pythonVersionPattern.FindStringSubmatch(output)

	if len(matches) != 2 {
		return "", fmt.Errorf("could not parse python version from %q", strings.TrimSpace(output))
	}

	return matches[1], nil
}

func getActivePythonVersion(ctx context.Context) (string, error) {
	for _, binary := range []string{"python3", "python"} {
		if _, err := exec.LookPath(binary); err != nil {
			continue
		}

		cmd := exec.CommandContext(ctx, binary, "--version")

		out, err := cmd.CombinedOutput()
		if err != nil {
			continue
		}

		return parsePythonVersion(string(out))
	}

	return "", fmt.Errorf("python not found")
}

func getActivePythonPath() (string, error) {
	if path, err := exec.LookPath("python3"); err == nil {
		return path, nil
	}

	return exec.LookPath("python")
}

// InspectPython detects every installed python version manager, then collects
// the active python binary and the versions each manager has installed.
func InspectPython(ctx context.Context) PythonInfo {
	info := PythonInfo{}

	if version, err := getActivePythonVersion(ctx); err == nil {
		info.ActiveVersion = version
	}

	if path, err := getActivePythonPath(); err == nil {
		info.ActivePath = path
	}

	for _, manager := range detectPythonManagers() {
		var versions []string
		var err error

		switch manager {
		case "mise":
			versions, err = getMisePythonVersions(ctx)

		case "pyenv":
			versions, err = getPyenvPythonVersions(ctx)

		case "asdf":
			versions, err = getAsdfPythonVersions(ctx)

		case "uv":
			versions, err = getUvPythonVersions(ctx)
		}

		if err != nil {
			continue
		}

		info.Managers = append(info.Managers, PythonManagerResult{
			Manager:  manager,
			Versions: versions,
		})
	}

	return info
}

// PythonCheck reports the active python version and the versions installed by
// every detected version manager.
type PythonCheck struct{}

func (PythonCheck) Name() string {
	return "Python"
}

func (PythonCheck) Run(ctx context.Context) model.Result {
	info := InspectPython(ctx)

	return runtimeResult("Python", "python", info.ActiveVersion, info.ActivePath, info.Managers)
}
