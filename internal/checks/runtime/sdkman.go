package runtime

import (
	"os"
	"path/filepath"
)

func sdkmanDir() (string, error) {
	dir := os.Getenv("SDKMAN_DIR")

	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		dir = filepath.Join(home, ".sdkman")
	}

	return dir, nil
}

// hasSDKMAN checks the sdkman home instead of exec.LookPath: sdk is a shell
// function, so there is usually no executable named "sdk" on PATH.
func hasSDKMAN() bool {
	dir, err := sdkmanDir()
	if err != nil {
		return false
	}

	_, err = os.Stat(filepath.Join(dir, "bin", "sdkman-init.sh"))

	return err == nil
}

// getSdkmanJavaVersions lists the installed JDKs directly, since `sdk list
// java` prints every version in the registry instead of only the installed
// ones. SDKMAN keeps each installed version under candidates/java, with a
// "current" symlink pointing at the active one.
func getSdkmanJavaVersions() ([]string, error) {
	dir, err := sdkmanDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(filepath.Join(dir, "candidates", "java"))
	if err != nil {
		return nil, err
	}

	var versions []string

	for _, entry := range entries {
		name := entry.Name()

		if name == "current" || !looksLikeVersion(name) {
			continue
		}

		versions = append(versions, name)
	}

	return dedupeVersions(versions), nil
}
