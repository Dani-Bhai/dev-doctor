package runtime

import (
	"slices"
	"testing"
)

func TestParsePythonVersion(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "python 3",
			output: "Python 3.11.15\n",
			want:   "3.11.15",
		},
		{
			name:   "python 2",
			output: "Python 2.7.18\n",
			want:   "2.7.18",
		},
		{
			name:   "unparseable",
			output: "python: command not found\n",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePythonVersion(tt.output)

			if tt.want == "" {
				if err == nil {
					t.Fatalf("parsePythonVersion() expected an error, got %q", got)
				}

				return
			}

			if err != nil {
				t.Fatalf("parsePythonVersion() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("parsePythonVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParsePyenvVersions(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name: "installed versions and active marker",
			output: `  system
  3.10.13
  3.11.9
* 3.12.1 (set by /home/user/.pyenv/version)
`,
			want: []string{"3.10.13", "3.11.9", "3.12.1"},
		},
		{
			name:   "system only",
			output: "* system (set by /home/user/.pyenv/version)\n",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePyenvVersions(tt.output)

			if !slices.Equal(got, tt.want) {
				t.Errorf("parsePyenvVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseUvPythonVersions(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name: "managed interpreters",
			output: `cpython-3.12.1-linux-x86_64-gnu    /home/user/.local/share/uv/python/cpython-3.12.1-linux-x86_64-gnu/bin/python3.12
cpython-3.12.1-linux-x86_64-gnu    /home/user/.local/share/uv/python/cpython-3.12.1-linux-x86_64-gnu/bin/python3
cpython-3.11.9-linux-x86_64-gnu    /home/user/.local/share/uv/python/cpython-3.11.9-linux-x86_64-gnu/bin/python3.11
`,
			want: []string{"cpython-3.12.1-linux-x86_64-gnu", "cpython-3.11.9-linux-x86_64-gnu"},
		},
		{
			name:   "nothing managed",
			output: "",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseUvPythonVersions(tt.output)

			if !slices.Equal(got, tt.want) {
				t.Errorf("parseUvPythonVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}
