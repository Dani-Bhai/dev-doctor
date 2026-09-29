package runtime

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestParseJavaVersion(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name: "openjdk",
			output: `openjdk version "17.0.2" 2022-01-18
OpenJDK Runtime Environment (build 17.0.2+8-86)
OpenJDK 64-Bit Server VM (build 17.0.2+8-86, mixed mode, sharing)
`,
			want: "17.0.2",
		},
		{
			name:   "java 8",
			output: "java version \"1.8.0_401\"\n",
			want:   "1.8.0_401",
		},
		{
			name:   "unparseable",
			output: "java: command not found\n",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseJavaVersion(tt.output)

			if tt.want == "" {
				if err == nil {
					t.Fatalf("parseJavaVersion() expected an error, got %q", got)
				}

				return
			}

			if err != nil {
				t.Fatalf("parseJavaVersion() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("parseJavaVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseJenvVersions(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name: "aliases and active marker",
			output: `* system (set by /home/user/.jenv/version)
  1.8
  1.8.0.232
  21.0
  21.0.2
  openjdk64-21.0.2
* openjdk64-17.0.8 (set by /home/user/.jenv/version)
`,
			want: []string{
				"1.8",
				"1.8.0.232",
				"21.0",
				"21.0.2",
				"openjdk64-21.0.2",
				"openjdk64-17.0.8",
			},
		},
		{
			name:   "system only",
			output: "* system (set by /home/user/.jenv/version)\n",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseJenvVersions(tt.output)

			if !slices.Equal(got, tt.want) {
				t.Errorf("parseJenvVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasSDKMAN(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SDKMAN_DIR", dir)

	if hasSDKMAN() {
		t.Fatal("hasSDKMAN() = true without sdkman-init.sh")
	}

	initDir := filepath.Join(dir, "bin")

	if err := os.MkdirAll(initDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(initDir, "sdkman-init.sh"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	if !hasSDKMAN() {
		t.Error("hasSDKMAN() = false with sdkman-init.sh present")
	}
}

func TestGetSdkmanJavaVersions(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SDKMAN_DIR", dir)

	for _, name := range []string{"17.0.2-tem", "21.0.10-tem"} {
		if err := os.MkdirAll(filepath.Join(dir, "candidates", "java", name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// The current symlink points at the active version and is not a version.
	if err := os.Symlink(
		filepath.Join(dir, "candidates", "java", "17.0.2-tem"),
		filepath.Join(dir, "candidates", "java", "current"),
	); err != nil {
		t.Fatal(err)
	}

	got, err := getSdkmanJavaVersions()
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"17.0.2-tem", "21.0.10-tem"}

	if !slices.Equal(got, want) {
		t.Errorf("getSdkmanJavaVersions() = %v, want %v", got, want)
	}
}
