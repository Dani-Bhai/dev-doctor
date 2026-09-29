package docker

import (
	"errors"
	"slices"
	"testing"
)

func TestParseClientVersion(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "standard output",
			output: "Docker version 29.7.2, build a7dcaa6fdb\n",
			want:   "29.7.2",
		},
		{
			name:   "different build",
			output: "Docker version 20.10.24, build 297e128\n",
			want:   "20.10.24",
		},
		{
			name:   "unexpected output",
			output: "docker: command not found\n",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseClientVersion(tt.output); got != tt.want {
				t.Errorf("parseClientVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseDaemonInfo(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    daemonInfo
		wantErr bool
	}{
		{
			name:   "full output",
			output: "29.7.2|Arch Linux|overlayfs|4|9|21\n",
			want: daemonInfo{
				serverVersion:     "29.7.2",
				operatingSystem:   "Arch Linux",
				driver:            "overlayfs",
				containersRunning: 4,
				containers:        9,
				images:            21,
			},
		},
		{
			name:   "empty counts",
			output: "29.7.2|Debian GNU/Linux 12|overlay2|||\n",
			want: daemonInfo{
				serverVersion:   "29.7.2",
				operatingSystem: "Debian GNU/Linux 12",
				driver:          "overlay2",
			},
		},
		{
			name:    "unexpected fields",
			output:  "29.7.2|Arch Linux\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDaemonInfo(tt.output)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseDaemonInfo() expected error, got %+v", got)
				}

				return
			}

			if err != nil {
				t.Fatalf("parseDaemonInfo() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("parseDaemonInfo() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseUsage(t *testing.T) {
	output := `Images|21|28.87GB|11.31GB (39%)
Containers|9|94.64MB|4.096kB (0%)
Local Volumes|37|2.056GB|726.9MB (35%)
Build Cache|44|12.78GB|0B
`

	want := []DiskUsage{
		{Type: "Images", Count: "21", Size: "28.87GB", Reclaimable: "11.31GB (39%)"},
		{Type: "Containers", Count: "9", Size: "94.64MB", Reclaimable: "4.096kB (0%)"},
		{Type: "Local Volumes", Count: "37", Size: "2.056GB", Reclaimable: "726.9MB (35%)"},
		{Type: "Build Cache", Count: "44", Size: "12.78GB", Reclaimable: "0B"},
	}

	got := parseUsage(output)

	if !slices.Equal(got, want) {
		t.Errorf("parseUsage() = %v, want %v", got, want)
	}
}

func TestParseContainers(t *testing.T) {
	output := `qa-backend|robocalling-backend:qa|Up 2 minutes
qa-postgres|postgres:17|Up 2 minutes (healthy)
qa-redis|redis/redis-stack:latest|Up 2 minutes (healthy)
`

	want := []Container{
		{Name: "qa-backend", Image: "robocalling-backend:qa", Status: "Up 2 minutes"},
		{Name: "qa-postgres", Image: "postgres:17", Status: "Up 2 minutes (healthy)"},
		{Name: "qa-redis", Image: "redis/redis-stack:latest", Status: "Up 2 minutes (healthy)"},
	}

	got := parseContainers(output)

	if !slices.Equal(got, want) {
		t.Errorf("parseContainers() = %v, want %v", got, want)
	}
}

func TestDaemonErrorMessage(t *testing.T) {
	tests := []struct {
		name   string
		stdout string
		stderr string
		want   string
	}{
		{
			name:   "prefers stderr over placeholder stdout",
			stdout: "||||0|0|0\n",
			stderr: "Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?\n",
			want:   "Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?",
		},
		{
			name:   "falls back to stdout",
			stdout: "some docker failure\n",
			stderr: "\n",
			want:   "some docker failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := daemonErrorMessage(tt.stdout, tt.stderr, errors.New("exit status 1"))

			if got != tt.want {
				t.Errorf("daemonErrorMessage() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("falls back to error", func(t *testing.T) {
		got := daemonErrorMessage("", "", errors.New("exit status 1"))

		if got != "exit status 1" {
			t.Errorf("daemonErrorMessage() = %q, want %q", got, "exit status 1")
		}
	})
}

func TestDockerSummary(t *testing.T) {
	tests := []struct {
		name string
		info Info
		want string
	}{
		{
			name: "daemon not running",
			info: Info{Version: "29.7.2"},
			want: "",
		},
		{
			name: "daemon running",
			info: Info{
				Version:       "29.7.2",
				DaemonRunning: true,
				RunningCount:  4,
				TotalCount:    9,
				ImageCount:    21,
			},
			want: "  daemon running · 4 running / 9 total containers · 21 images",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dockerSummary(tt.info); got != tt.want {
				t.Errorf("dockerSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFirstNonEmptyLine(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "skips leading empty lines",
			output: "\n\nCannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?\n",
			want:   "Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?",
		},
		{
			name:   "empty output",
			output: "\n",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstNonEmptyLine(tt.output); got != tt.want {
				t.Errorf("firstNonEmptyLine() = %q, want %q", got, tt.want)
			}
		})
	}
}
