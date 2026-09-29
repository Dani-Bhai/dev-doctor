package runtime

import (
	"slices"
	"testing"
)

func TestParseGoVersion(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "release",
			output: "go version go1.27.1 linux/amd64\n",
			want:   "1.27.1",
		},
		{
			name:   "devel",
			output: "go version devel go1.28-abc123 linux/amd64\n",
			want:   "1.28-abc123",
		},
		{
			name:   "unparseable",
			output: "go: command not found\n",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGoVersion(tt.output)

			if tt.want == "" {
				if err == nil {
					t.Fatalf("parseGoVersion() expected an error, got %q", got)
				}

				return
			}

			if err != nil {
				t.Fatalf("parseGoVersion() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("parseGoVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseGoenvVersions(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name: "installed versions and active marker",
			output: `  system
  1.21.5
* 1.22.1 (set by /home/user/.goenv/version)
`,
			want: []string{"1.21.5", "1.22.1"},
		},
		{
			name:   "system only",
			output: "* system (set by /home/user/.goenv/version)\n",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGoenvVersions(tt.output)

			if !slices.Equal(got, tt.want) {
				t.Errorf("parseGoenvVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseGVMVersions(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name: "installed versions and default marker",
			output: `gvm gos (installed)

   go1.20.8
=> go1.22.4
   system
`,
			want: []string{"go1.20.8", "go1.22.4"},
		},
		{
			name: "nothing installed",
			output: `gvm gos (installed)

`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGVMVersions(tt.output)

			if !slices.Equal(got, tt.want) {
				t.Errorf("parseGVMVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}
