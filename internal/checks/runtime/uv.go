package runtime

import (
	"context"
	"os/exec"
	"strings"
)

// getUvPythonVersions only reports interpreters installed by uv itself, not
// the system or other managers' installations uv can also see.
func getUvPythonVersions(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "uv", "python", "list", "--only-installed", "--managed-python")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return parseUvPythonVersions(string(out)), nil
}

func parseUvPythonVersions(output string) []string {
	var versions []string

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)

		if len(fields) == 0 || !looksLikeVersion(fields[0]) {
			continue
		}

		versions = append(versions, fields[0])
	}

	return dedupeVersions(versions)
}
