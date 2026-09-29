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

type JavaInfo struct {
	ActiveVersion string
	ActivePath    string
	Managers      []JavaManagerResult
}

var javaVersionPattern = regexp.MustCompile(`"([^"]+)"`)

func detectJavaManagers() []string {
	var managers []string

	if system.CommandExists("mise") {
		managers = append(managers, "mise")
	}

	if hasSDKMAN() {
		managers = append(managers, "sdkman")
	}

	if system.CommandExists("asdf") {
		managers = append(managers, "asdf")
	}

	if system.CommandExists("jenv") {
		managers = append(managers, "jenv")
	}

	return managers
}

func parseJavaVersion(output string) (string, error) {
	matches := javaVersionPattern.FindStringSubmatch(output)

	if len(matches) != 2 {
		return "", fmt.Errorf("could not parse java version from %q", strings.TrimSpace(output))
	}

	return matches[1], nil
}

func getActiveJavaVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "java", "-version")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return parseJavaVersion(string(out))
}

func getActiveJavaPath() (string, error) {
	return exec.LookPath("java")
}

// InspectJava detects every installed java version manager, then collects the
// active java binary and the versions each manager has installed.
func InspectJava(ctx context.Context) JavaInfo {
	info := JavaInfo{}

	if version, err := getActiveJavaVersion(ctx); err == nil {
		info.ActiveVersion = version
	}

	if path, err := getActiveJavaPath(); err == nil {
		info.ActivePath = path
	}

	for _, manager := range detectJavaManagers() {
		var versions []string
		var err error

		switch manager {
		case "mise":
			versions, err = getMiseJavaVersions(ctx)

		case "sdkman":
			versions, err = getSdkmanJavaVersions()

		case "asdf":
			versions, err = getAsdfJavaVersions(ctx)

		case "jenv":
			versions, err = getJenvJavaVersions(ctx)
		}

		if err != nil {
			continue
		}

		info.Managers = append(info.Managers, JavaManagerResult{
			Manager:  manager,
			Versions: versions,
		})
	}

	return info
}

// JavaCheck reports the active java version and the versions installed by
// every detected version manager.
type JavaCheck struct{}

func (JavaCheck) Name() string {
	return "Java"
}

func (JavaCheck) Run(ctx context.Context) model.Result {
	info := InspectJava(ctx)

	return runtimeResult("Java", "java", info.ActiveVersion, info.ActivePath, info.Managers)
}
