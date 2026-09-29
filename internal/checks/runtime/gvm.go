package runtime

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func gvmDir() (string, error) {
	dir := os.Getenv("GVM_DIR")

	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		dir = filepath.Join(home, ".gvm")
	}

	return dir, nil
}

// hasGVM checks the gvm home instead of exec.LookPath: gvm is a shell
// function, so there is usually no executable named "gvm" on PATH.
func hasGVM() bool {
	dir, err := gvmDir()
	if err != nil {
		return false
	}

	_, err = os.Stat(filepath.Join(dir, "scripts", "gvm"))

	return err == nil
}

// getGVMGoVersions sources the gvm script in bash and runs `gvm list`,
// because gvm is a shell function and cannot be executed directly.
func getGVMGoVersions(ctx context.Context) ([]string, error) {
	dir, err := gvmDir()
	if err != nil {
		return nil, err
	}

	script := filepath.Join(dir, "scripts", "gvm")

	if _, err := os.Stat(script); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(
		ctx,
		"bash",
		"-c",
		`source "$1" && gvm list`,
		"bash",
		script,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return parseGVMVersions(string(out)), nil
}

func parseGVMVersions(output string) []string {
	var versions []string

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimSpace(strings.TrimPrefix(line, "=>"))

		if !looksLikeVersion(line) {
			continue
		}

		versions = append(versions, line)
	}

	return dedupeVersions(versions)
}
