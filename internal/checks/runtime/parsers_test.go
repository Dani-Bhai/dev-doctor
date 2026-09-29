package runtime

import (
	"slices"
	"testing"
)

func TestParseMiseVersions(t *testing.T) {
	tests := []struct {
		name   string
		tool   string
		output string
		want   []string
	}{
		{
			name:   "single installed version",
			tool:   "node",
			output: "node  26.2.0  ~/.config/mise/config.toml  latest\n",
			want:   []string{"26.2.0"},
		},
		{
			name: "multiple versions with header",
			tool: "node",
			output: "Tool  Version  Source                      Requested\n" +
				"node  20.19.5  ~/.tool-versions            20\n" +
				"node  24.8.0   ~/.config/mise/config.toml  latest\n",
			want: []string{"20.19.5", "24.8.0"},
		},
		{
			name:   "no tool column",
			tool:   "node",
			output: "26.2.0  ~/.config/mise/config.toml  latest\n",
			want:   []string{"26.2.0"},
		},
		{
			name: "java versions with vendor prefix",
			tool: "java",
			output: "Tool  Version         Source           Requested\n" +
				"java  temurin-17.0.2  ~/.tool-versions  17\n" +
				"java  temurin-21.0.2  mise.toml         latest\n",
			want: []string{"temurin-17.0.2", "temurin-21.0.2"},
		},
		{
			name: "python versions",
			tool: "python",
			output: "Tool    Version  Source           Requested\n" +
				"python  3.11.15  ~/.tool-versions  3.11\n" +
				"python  3.14.6   mise.toml         latest\n",
			want: []string{"3.11.15", "3.14.6"},
		},
		{
			name: "go versions",
			tool: "go",
			output: "Tool  Version  Source           Requested\n" +
				"go    1.27.1   ~/.tool-versions  latest\n",
			want: []string{"1.27.1"},
		},
		{
			name: "ignores warnings and headers",
			tool: "node",
			output: "mise WARN  mise version 2026.9.16 available\n" +
				"Tool  Version  Source  Requested\n",
			want: nil,
		},
		{
			name:   "dedupes the same version from two sources",
			tool:   "node",
			output: "node  24.8.0  ~/.tool-versions\nnode  24.8.0  mise.toml\n",
			want:   []string{"24.8.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMiseVersions(tt.output, tt.tool)

			if !slices.Equal(got, tt.want) {
				t.Errorf("parseMiseVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseNVMVersions(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name: "installed versions and aliases",
			output: `->     v24.8.0
       v20.19.5
       v18.20.4
default -> v24.8.0
iojs -> N/A (default)
unstable -> N/A (default)
node -> stable (-> v24.8.0) (default)
stable -> 24.8 (-> v24.8.0) (default)
lts/* -> lts/iron (-> v20.19.5)
`,
			want: []string{"v24.8.0", "v20.19.5", "v18.20.4"},
		},
		{
			name: "duplicate active entry",
			output: `-> v24.8.0
   v24.8.0
`,
			want: []string{"v24.8.0"},
		},
		{
			name:   "nothing installed",
			output: "N/A\n",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNVMVersions(tt.output)

			if !slices.Equal(got, tt.want) {
				t.Errorf("parseNVMVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAsdfVersions(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name: "active version marked with asterisk",
			output: `  18.20.4
* 20.19.5
  22.20.0
`,
			want: []string{"18.20.4", "20.19.5", "22.20.0"},
		},
		{
			name: "java versions with vendor prefix",
			output: `  17.0.2
* temurin-21.0.2
`,
			want: []string{"17.0.2", "temurin-21.0.2"},
		},
		{
			name: "python versions",
			output: `  3.11.15
* 3.12.1
`,
			want: []string{"3.11.15", "3.12.1"},
		},
		{
			name: "go versions",
			output: `  1.27.1
`,
			want: []string{"1.27.1"},
		},
		{
			name:   "nothing installed",
			output: "No versions installed\n",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAsdfVersions(tt.output)

			if !slices.Equal(got, tt.want) {
				t.Errorf("parseAsdfVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFnmVersions(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name: "versions and aliases",
			output: `* v20.19.5 default
* v18.20.4
  v22.20.0
`,
			want: []string{"v20.19.5", "v18.20.4", "v22.20.0"},
		},
		{
			name:   "system alias is not a version",
			output: "* system\n",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFnmVersions(tt.output)

			if !slices.Equal(got, tt.want) {
				t.Errorf("parseFnmVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}
