package runtime

import (
	"context"
	"os/exec"
	"strings"
)

func getFnmNodeVersions(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "fnm", "list")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return parseFnmVersions(string(out)), nil
}

func parseFnmVersions(output string) []string {
	var versions []string

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimSpace(strings.TrimPrefix(line, "*"))

		fields := strings.Fields(line)

		if len(fields) == 0 || !versionPattern.MatchString(fields[0]) {
			continue
		}

		versions = append(versions, fields[0])
	}

	return dedupeVersions(versions)
}
