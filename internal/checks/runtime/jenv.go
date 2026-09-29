package runtime

import (
	"context"
	"os/exec"
	"strings"
)

func getJenvJavaVersions(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "jenv", "versions")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return parseJenvVersions(string(out)), nil
}

// parseJenvVersions keeps every line, because jenv registers aliases next to
// the full JDK name.
func parseJenvVersions(output string) []string {
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
