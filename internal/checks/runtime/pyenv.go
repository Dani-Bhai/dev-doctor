package runtime

import (
	"context"
	"os/exec"
	"strings"
)

func getPyenvPythonVersions(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "pyenv", "versions")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return parsePyenvVersions(string(out)), nil
}

func parsePyenvVersions(output string) []string {
	var versions []string

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimSpace(strings.TrimPrefix(line, "*"))

		if index := strings.Index(line, " (set by "); index != -1 {
			line = line[:index]
		}

		line = strings.TrimSpace(line)

		if !looksLikeVersion(line) {
			continue
		}

		versions = append(versions, line)
	}

	return dedupeVersions(versions)
}
