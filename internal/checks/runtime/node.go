package runtime

import (
	"context"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Dani-Bhai/dev-doctor/internal/checks/system"
	"github.com/Dani-Bhai/dev-doctor/internal/model"
)

type NodeInfo struct {
	ActiveVersion string
	ActivePath    string
	Managers      []NodeManagerResult
}

var versionPattern = regexp.MustCompile(`^v?\d+(\.\d+)*$`)

// looksLikeVersion also accepts vendor prefixed versions such as
// temurin-17.0.2 or 17.0.2-tem, as long as they contain a digit.
var versionTokenPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]*$`)

func looksLikeVersion(token string) bool {
	return versionTokenPattern.MatchString(token) && strings.ContainsAny(token, "0123456789")
}

func detectNodeManagers() []string {
	var managers []string

	if system.CommandExists("mise") {
		managers = append(managers, "mise")
	}

	if hasNVM() {
		managers = append(managers, "nvm")
	}

	if system.CommandExists("asdf") {
		managers = append(managers, "asdf")
	}

	if system.CommandExists("fnm") {
		managers = append(managers, "fnm")
	}

	return managers
}

func getActiveNodeVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "node", "--version")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func getActiveNodePath() (string, error) {
	return exec.LookPath("node")
}

// InspectNode detects every installed node version manager, then collects the
// active node binary and the versions each manager has installed.
func InspectNode(ctx context.Context) NodeInfo {
	info := NodeInfo{}

	if version, err := getActiveNodeVersion(ctx); err == nil {
		info.ActiveVersion = version
	}

	if path, err := getActiveNodePath(); err == nil {
		info.ActivePath = path
	}

	for _, manager := range detectNodeManagers() {
		var versions []string
		var err error

		switch manager {
		case "mise":
			versions, err = getMiseNodeVersions(ctx)

		case "nvm":
			versions, err = getNVMNodeVersions(ctx)

		case "asdf":
			versions, err = getAsdfNodeVersions(ctx)

		case "fnm":
			versions, err = getFnmNodeVersions(ctx)
		}

		if err != nil {
			continue
		}

		info.Managers = append(info.Managers, NodeManagerResult{
			Manager:  manager,
			Versions: versions,
		})
	}

	return info
}

// NodeCheck reports the active node version and the versions installed by
// every detected version manager.
type NodeCheck struct{}

func (NodeCheck) Name() string {
	return "Node.js"
}

func (NodeCheck) Run(ctx context.Context) model.Result {
	info := InspectNode(ctx)

	return runtimeResult("Node.js", "node", info.ActiveVersion, info.ActivePath, info.Managers)
}

func dedupeVersions(versions []string) []string {
	seen := make(map[string]struct{}, len(versions))
	unique := make([]string, 0, len(versions))

	for _, version := range versions {
		if _, ok := seen[version]; ok {
			continue
		}

		seen[version] = struct{}{}
		unique = append(unique, version)
	}

	return unique
}
