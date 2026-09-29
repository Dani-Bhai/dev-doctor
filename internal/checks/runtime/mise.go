package runtime

import (
	"context"
	"os/exec"
	"strings"
)

func getMiseNodeVersions(ctx context.Context) ([]string, error) {
	return getMiseVersions(ctx, "node")
}

func getMiseJavaVersions(ctx context.Context) ([]string, error) {
	return getMiseVersions(ctx, "java")
}

func getMisePythonVersions(ctx context.Context) ([]string, error) {
	return getMiseVersions(ctx, "python")
}

func getMiseGoVersions(ctx context.Context) ([]string, error) {
	return getMiseVersions(ctx, "go")
}

func getMiseVersions(ctx context.Context, tool string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "mise", "ls", "--installed", tool)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return parseMiseVersions(string(out), tool), nil
}

// parseMiseVersions accepts a bare version in the first column too, because
// older mise versions omit the Tool column when a tool is passed.
func parseMiseVersions(output, tool string) []string {
	var versions []string

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)

		if len(fields) == 0 {
			continue
		}

		var version string

		switch {
		case fields[0] == tool && len(fields) > 1:
			version = fields[1]

		case looksLikeVersion(fields[0]):
			version = fields[0]

		default:
			continue
		}

		if !looksLikeVersion(version) {
			continue
		}

		versions = append(versions, version)
	}

	return dedupeVersions(versions)
}
