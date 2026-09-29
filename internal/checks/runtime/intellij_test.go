package runtime

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

const testJDKTable = `<application>
  <component name="ProjectJdkTable">
    <jdk version="2">
      <name value="26" />
      <type value="JavaSDK" />
      <version value="Oracle OpenJDK 26.0.1" />
      <homePath value="$USER_HOME$/.jdks/oracle-26" />
    </jdk>
    <jdk version="2">
      <name value="17" />
      <version value="Java 17.0.20" />
      <homePath value="$USER_HOME$/.jdks/temurin-17" />
    </jdk>
  </component>
</application>
`

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeTestJDKTable(t *testing.T, home, ide, content string) {
	t.Helper()

	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	path := filepath.Join(home, ".config", "JetBrains", ide, "options", "jdk.table.xml")

	writeTestFile(t, path, content)
}

func writeTestJDK(t *testing.T, dir, release string) {
	t.Helper()

	writeTestFile(t, filepath.Join(dir, "bin", "java"), "")

	if release != "" {
		writeTestFile(t, filepath.Join(dir, "release"), "JAVA_VERSION=\""+release+"\"\n")
	}
}

func TestReadJDKTable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jdk.table.xml")

	writeTestFile(t, path, testJDKTable)

	entries := readJDKTable(path)

	if len(entries) != 2 {
		t.Fatalf("readJDKTable() returned %d entries, want 2", len(entries))
	}

	if got := entries[0].Name.Value; got != "26" {
		t.Errorf("first entry name = %q, want %q", got, "26")
	}

	if got := entries[0].Version.Value; got != "Oracle OpenJDK 26.0.1" {
		t.Errorf("first entry version = %q, want %q", got, "Oracle OpenJDK 26.0.1")
	}

	if got := entries[1].HomePath.Value; got != "$USER_HOME$/.jdks/temurin-17" {
		t.Errorf("second entry home = %q, want %q", got, "$USER_HOME$/.jdks/temurin-17")
	}

	t.Run("missing file", func(t *testing.T) {
		if got := readJDKTable(filepath.Join(t.TempDir(), "missing.xml")); got != nil {
			t.Errorf("readJDKTable() = %v, want nil", got)
		}
	})

	t.Run("malformed xml", func(t *testing.T) {
		bad := filepath.Join(t.TempDir(), "bad.xml")

		writeTestFile(t, bad, "<application><component>")

		if got := readJDKTable(bad); got != nil {
			t.Errorf("readJDKTable() = %v, want nil", got)
		}
	})
}

func TestReadJavaReleaseVersion(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "jdk")

	writeTestFile(t, filepath.Join(dir, "release"), "IMPLEMENTOR=\"JetBrains s.r.o.\"\nJAVA_VERSION=\"25.0.3\"\n")

	if got := readJavaReleaseVersion(dir); got != "25.0.3" {
		t.Errorf("readJavaReleaseVersion() = %q, want %q", got, "25.0.3")
	}

	t.Run("missing release file", func(t *testing.T) {
		if got := readJavaReleaseVersion(t.TempDir()); got != "" {
			t.Errorf("readJavaReleaseVersion() = %q, want empty", got)
		}
	})
}

func TestCollectIntelliJJavaVersions(t *testing.T) {
	t.Run("jdk table and downloads without duplicates", func(t *testing.T) {
		home := t.TempDir()

		writeTestJDKTable(t, home, "IntelliJIdea2026.2", testJDKTable)

		// Referenced by the table and must not be listed twice.
		writeTestJDK(t, filepath.Join(home, ".jdks", "oracle-26"), "26.0.1")

		// Downloaded by the IDE but not yet in the table.
		writeTestJDK(t, filepath.Join(home, ".jdks", "temurin-21.0.5"), "21.0.5")

		// Incomplete directory without bin/java.
		if err := os.MkdirAll(filepath.Join(home, ".jdks", "incomplete"), 0o755); err != nil {
			t.Fatal(err)
		}

		got := collectIntelliJJavaVersions(home, nil)

		want := []string{
			"26 (Oracle OpenJDK 26.0.1)",
			"17 (Java 17.0.20)",
			"temurin-21.0.5 (21.0.5)",
		}

		if !slices.Equal(got, want) {
			t.Errorf("collectIntelliJJavaVersions() = %v, want %v", got, want)
		}
	})

	t.Run("bundled JetBrains Runtime", func(t *testing.T) {
		home := t.TempDir()

		t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

		jbr := filepath.Join(home, ".local", "share", "JetBrains", "Toolbox", "apps", "intellij-idea-ultimate", "ch-0", "251.1234", "jbr")

		writeTestJDK(t, jbr, "25.0.3")

		patterns := []string{
			filepath.Join(home, ".local", "share", "JetBrains", "Toolbox", "apps", "*", "*", "*", "jbr"),
		}

		got := collectIntelliJJavaVersions(home, patterns)

		want := []string{"jbr (25.0.3)"}

		if !slices.Equal(got, want) {
			t.Errorf("collectIntelliJJavaVersions() = %v, want %v", got, want)
		}
	})

	t.Run("nothing installed", func(t *testing.T) {
		home := t.TempDir()

		t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

		if got := collectIntelliJJavaVersions(home, nil); got != nil {
			t.Errorf("collectIntelliJJavaVersions() = %v, want nil", got)
		}
	})
}

func TestHasIntelliJAt(t *testing.T) {
	home := t.TempDir()

	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	if hasIntelliJAt(home, nil) {
		t.Fatal("hasIntelliJAt() = true for an empty home")
	}

	writeTestJDKTable(t, home, "IntelliJIdea2026.2", testJDKTable)

	if !hasIntelliJAt(home, nil) {
		t.Error("hasIntelliJAt() = false with a jdk.table.xml present")
	}
}
