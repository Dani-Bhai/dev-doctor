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

type GoInfo struct {
	ActiveVersion string
	ActivePath    string
	Managers      []GoManagerResult
}

var goVersionPattern = regexp.MustCompile(`go(1\.[^\s]+)`)

func detectGoManagers() []string {
	var managers []string

	if system.CommandExists("mise") {
		managers = append(managers, "mise")
	}

	if hasGVM() {
		managers = append(managers, "gvm")
	}

	if system.CommandExists("asdf") {
		managers = append(managers, "asdf")
	}

	if system.CommandExists("goenv") {
		managers = append(managers, "goenv")
	}

	return managers
}

func parseGoVersion(output string) (string, error) {
	matches := goVersionPattern.FindStringSubmatch(output)

	if len(matches) != 2 {
		return "", fmt.Errorf("could not parse go version from %q", strings.TrimSpace(output))
	}

	return matches[1], nil
}

func getActiveGoVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "version")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return parseGoVersion(string(out))
}

func getActiveGoPath() (string, error) {
	return exec.LookPath("go")
}

// InspectGo detects every installed go version manager, then collects the
// active go binary and the versions each manager has installed.
func InspectGo(ctx context.Context) GoInfo {
	info := GoInfo{}

	if version, err := getActiveGoVersion(ctx); err == nil {
		info.ActiveVersion = version
	}

	if path, err := getActiveGoPath(); err == nil {
		info.ActivePath = path
	}

	for _, manager := range detectGoManagers() {
		var versions []string
		var err error

		switch manager {
		case "mise":
			versions, err = getMiseGoVersions(ctx)

		case "gvm":
			versions, err = getGVMGoVersions(ctx)

		case "asdf":
			versions, err = getAsdfGoVersions(ctx)

		case "goenv":
			versions, err = getGoenvGoVersions(ctx)
		}

		if err != nil {
			continue
		}

		info.Managers = append(info.Managers, GoManagerResult{
			Manager:  manager,
			Versions: versions,
		})
	}

	return info
}

// GoCheck reports the active go version and the versions installed by every
// detected version manager.
type GoCheck struct{}

func (GoCheck) Name() string {
	return "Go"
}

func (GoCheck) Run(ctx context.Context) model.Result {
	info := InspectGo(ctx)

	return runtimeResult("Go", "go", info.ActiveVersion, info.ActivePath, info.Managers)
}
