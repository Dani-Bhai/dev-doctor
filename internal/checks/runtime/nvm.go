package runtime

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var nvmVersionPattern = regexp.MustCompile(`^\s*(?:->\s*)?(v?\d+\.\d+\.\d+)\s*$`)

// hasNVM checks the nvm home instead of exec.LookPath: nvm is a shell
// function, so there is usually no executable named "nvm" on PATH.
func hasNVM() bool {
	if os.Getenv("NVM_DIR") != "" {
		return true
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	_, err = os.Stat(filepath.Join(home, ".nvm"))

	return err == nil
}

func nvmScriptPath() (string, error) {
	dir := os.Getenv("NVM_DIR")

	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		dir = filepath.Join(home, ".nvm")
	}

	script := filepath.Join(dir, "nvm.sh")

	if _, err := os.Stat(script); err != nil {
		return "", err
	}

	return script, nil
}

// getNVMNodeVersions sources nvm.sh in bash and runs `nvm ls`, because nvm is
// a shell function and cannot be executed directly.
func getNVMNodeVersions(ctx context.Context) ([]string, error) {
	script, err := nvmScriptPath()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(
		ctx,
		"bash",
		"-c",
		`source "$1" && nvm ls --no-colors`,
		"bash",
		script,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return parseNVMVersions(string(out)), nil
}

func parseNVMVersions(output string) []string {
	var versions []string

	for _, line := range strings.Split(output, "\n") {
		matches := nvmVersionPattern.FindStringSubmatch(line)

		if len(matches) != 2 {
			continue
		}

		versions = append(versions, matches[1])
	}

	return dedupeVersions(versions)
}
