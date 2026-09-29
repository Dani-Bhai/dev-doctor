package runtime

import (
	"context"
	"os/exec"
	"strings"
)

func getAsdfNodeVersions(ctx context.Context) ([]string, error) {
	return getAsdfVersions(ctx, "nodejs")
}

func getAsdfJavaVersions(ctx context.Context) ([]string, error) {
	return getAsdfVersions(ctx, "java")
}

func getAsdfPythonVersions(ctx context.Context) ([]string, error) {
	return getAsdfVersions(ctx, "python")
}

func getAsdfGoVersions(ctx context.Context) ([]string, error) {
	return getAsdfVersions(ctx, "golang")
}

func getAsdfVersions(ctx context.Context, plugin string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "asdf", "list", plugin)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return parseAsdfVersions(string(out)), nil
}

func parseAsdfVersions(output string) []string {
	var versions []string

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimSpace(strings.TrimPrefix(line, "*"))

		if !looksLikeVersion(line) {
			continue
		}

		versions = append(versions, line)
	}

	return dedupeVersions(versions)
}
