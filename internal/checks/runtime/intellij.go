package runtime

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const intellijManager = "IntelliJ"

var javaReleaseVersionPattern = regexp.MustCompile(`JAVA_VERSION="([^"]+)"`)

type jdkTableValue struct {
	Value string `xml:"value,attr"`
}

type jdkTableEntry struct {
	Name     jdkTableValue `xml:"name"`
	Version  jdkTableValue `xml:"version"`
	HomePath jdkTableValue `xml:"homePath"`
}

type jdkTable struct {
	Components []struct {
		Name string          `xml:"name,attr"`
		JDKs []jdkTableEntry `xml:"jdk"`
	} `xml:"component"`
}

// readJDKTable returns the JDKs configured in an IntelliJ jdk.table.xml file.
func readJDKTable(path string) []jdkTableEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var table jdkTable

	if err := xml.Unmarshal(data, &table); err != nil {
		return nil
	}

	for _, component := range table.Components {
		if component.Name == "ProjectJdkTable" {
			return component.JDKs
		}
	}

	return nil
}

// jetbrainsConfigDir supports both the default ~/.config/JetBrains and a
// custom XDG_CONFIG_HOME.
func jetbrainsConfigDir(home string) string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "JetBrains")
	}

	return filepath.Join(home, ".config", "JetBrains")
}

func intellijJDKTablePaths(home string) []string {
	pattern := filepath.Join(jetbrainsConfigDir(home), "*", "options", "jdk.table.xml")

	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}

	return paths
}

// intellijJBRPatterns covers the bundled JetBrains Runtime of Toolbox, snap
// and system installs.
func intellijJBRPatterns(home string) []string {
	return []string{
		filepath.Join(home, ".local", "share", "JetBrains", "Toolbox", "apps", "*", "*", "*", "jbr"),
		filepath.Join(home, "idea*", "jbr"),
		"/opt/idea*/jbr",
		"/opt/intellij*/jbr",
		"/usr/lib/intellij*/jbr",
		"/snap/intellij-idea*/current/jbr",
	}
}

func findJBRs(patterns []string) []string {
	var runtimes []string

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, match := range matches {
			if hasJavaBinary(match) {
				runtimes = append(runtimes, match)
			}
		}
	}

	return runtimes
}

func hasJavaBinary(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "bin", "java"))

	return err == nil
}

func readJavaReleaseVersion(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "release"))
	if err != nil {
		return ""
	}

	matches := javaReleaseVersionPattern.FindStringSubmatch(string(data))

	if len(matches) != 2 {
		return ""
	}

	return matches[1]
}

// jdkDisplayName formats a JDK directory the way the IDE labels it:
// "temurin-21.0.5 (21.0.5)".
func jdkDisplayName(name, dir string) string {
	version := readJavaReleaseVersion(dir)

	if version == "" {
		return name
	}

	return fmt.Sprintf("%s (%s)", name, version)
}

func hasIntelliJAt(home string, jbrPatterns []string) bool {
	if len(intellijJDKTablePaths(home)) > 0 {
		return true
	}

	if entries, err := os.ReadDir(filepath.Join(home, ".jdks")); err == nil && len(entries) > 0 {
		return true
	}

	return len(findJBRs(jbrPatterns)) > 0
}

// collectIntelliJJavaVersions lists every runtime IntelliJ knows about: the
// JDKs in its jdk.table.xml, the ones it downloaded into ~/.jdks and the
// bundled JetBrains Runtime. Entries are deduplicated by home directory.
func collectIntelliJJavaVersions(home string, jbrPatterns []string) []string {
	var versions []string

	seen := make(map[string]struct{})

	add := func(key, display string) {
		if key == "" || display == "" {
			return
		}

		if _, ok := seen[key]; ok {
			return
		}

		seen[key] = struct{}{}
		versions = append(versions, display)
	}

	for _, path := range intellijJDKTablePaths(home) {
		for _, entry := range readJDKTable(path) {
			name := entry.Name.Value
			if name == "" {
				continue
			}

			key := strings.ReplaceAll(entry.HomePath.Value, "$USER_HOME$", home)
			if key == "" {
				key = "name:" + name
			} else {
				key = filepath.Clean(key)
			}

			display := name

			if entry.Version.Value != "" && entry.Version.Value != name {
				display = fmt.Sprintf("%s (%s)", name, entry.Version.Value)
			}

			add(key, display)
		}
	}

	jdksDir := filepath.Join(home, ".jdks")

	if entries, err := os.ReadDir(jdksDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			dir := filepath.Join(jdksDir, entry.Name())

			if !hasJavaBinary(dir) {
				continue
			}

			add(dir, jdkDisplayName(entry.Name(), dir))
		}
	}

	for _, jbr := range findJBRs(jbrPatterns) {
		add(jbr, jdkDisplayName("jbr", jbr))
	}

	return versions
}

func hasIntelliJ() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	return hasIntelliJAt(home, intellijJBRPatterns(home))
}

func getIntelliJJavaVersions() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	versions := collectIntelliJJavaVersions(home, intellijJBRPatterns(home))

	if len(versions) == 0 {
		return nil, fmt.Errorf("no IntelliJ runtimes found")
	}

	return versions, nil
}
