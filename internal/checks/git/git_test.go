package git

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestParseGitVersion(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "standard output",
			output: "git version 2.55.0\n",
			want:   "2.55.0",
		},
		{
			name:   "apple build suffix",
			output: "git version 2.39.3 (Apple Git-145)\n",
			want:   "2.39.3",
		},
		{
			name:   "unexpected output",
			output: "git: command not found\n",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseGitVersion(tt.output); got != tt.want {
				t.Errorf("parseGitVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseGhAccounts(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []Account
	}{
		{
			name: "single active account",
			output: `github.com
  ✓ Logged in to github.com account Dani-Bhai (keyring)
  - Active account: true
  - Git operations protocol: https
  - Token: gho_************************************
`,
			want: []Account{
				{Host: "github.com", Username: "Dani-Bhai", Source: "gh", Active: true},
			},
		},
		{
			name: "two accounts one active",
			output: `github.com
  ✓ Logged in to github.com account work-user (keyring)
  - Active account: false
  ✓ Logged in to github.com account Dani-Bhai (keyring)
  - Active account: true
`,
			want: []Account{
				{Host: "github.com", Username: "work-user", Source: "gh"},
				{Host: "github.com", Username: "Dani-Bhai", Source: "gh", Active: true},
			},
		},
		{
			name: "enterprise host",
			output: `github.example.com
  ✓ Logged in to github.example.com account dev (keyring)
  - Active account: true
`,
			want: []Account{
				{Host: "github.example.com", Username: "dev", Source: "gh", Active: true},
			},
		},
		{
			name:   "not logged in",
			output: "You are not logged into any GitHub hosts. Run gh auth login to authenticate.\n",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGhAccounts(tt.output)

			if !slices.Equal(got, tt.want) {
				t.Errorf("parseGhAccounts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseGlabAccounts(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []Account
	}{
		{
			name: "single account",
			output: `gitlab.com
  ✓ Logged in to gitlab.com as devdaniyal (/home/dani/.config/glab-cli/config.yml)
`,
			want: []Account{
				{Host: "gitlab.com", Username: "devdaniyal", Source: "glab"},
			},
		},
		{
			name: "self hosted",
			output: `gitlab.example.com
  ✓ Logged in to gitlab.example.com as dev (/home/dani/.config/glab-cli/config.yml)
`,
			want: []Account{
				{Host: "gitlab.example.com", Username: "dev", Source: "glab"},
			},
		},
		{
			name:   "not logged in",
			output: "No hosts configured\n",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGlabAccounts(tt.output)

			if !slices.Equal(got, tt.want) {
				t.Errorf("parseGlabAccounts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadCredentialStore(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []Account
	}{
		{
			name:    "username and token",
			content: "https://Dani-Bhai:ghp_secret@github.com\n",
			want: []Account{
				{Host: "github.com", Username: "Dani-Bhai", Source: "credentials"},
			},
		},
		{
			name:    "token only",
			content: "https://ghp_secret@github.com\n",
			want: []Account{
				{Host: "github.com", Username: "ghp_secret", Source: "credentials"},
			},
		},
		{
			name:    "multiple hosts and blank lines",
			content: "https://dani:token@github.com\n\nhttps://dani:token@gitlab.com\n",
			want: []Account{
				{Host: "github.com", Username: "dani", Source: "credentials"},
				{Host: "gitlab.com", Username: "dani", Source: "credentials"},
			},
		},
		{
			name:    "malformed lines are skipped",
			content: "not a url\nhttps://dani:token@github.com\n",
			want: []Account{
				{Host: "github.com", Username: "dani", Source: "credentials"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), ".git-credentials")

			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("writing credential store: %v", err)
			}

			got := readCredentialStore(path)

			if !slices.Equal(got, tt.want) {
				t.Errorf("readCredentialStore() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("missing file", func(t *testing.T) {
		if got := readCredentialStore(filepath.Join(t.TempDir(), "missing")); got != nil {
			t.Errorf("readCredentialStore() = %v, want nil", got)
		}
	})
}

func TestDedupeAccounts(t *testing.T) {
	accounts := []Account{
		{Host: "github.com", Username: "dani", Source: "gh"},
		{Host: "github.com", Username: "dani", Source: "gh"},
		{Host: "github.com", Username: "dani", Source: "credentials"},
	}

	want := []Account{
		{Host: "github.com", Username: "dani", Source: "gh"},
		{Host: "github.com", Username: "dani", Source: "credentials"},
	}

	if got := dedupeAccounts(accounts); !slices.Equal(got, want) {
		t.Errorf("dedupeAccounts() = %v, want %v", got, want)
	}
}

func TestIdentityString(t *testing.T) {
	tests := []struct {
		name string
		info Info
		want string
	}{
		{
			name: "name and email",
			info: Info{Name: "Dani", Email: "dani@example.com"},
			want: "Dani <dani@example.com>",
		},
		{
			name: "name only",
			info: Info{Name: "Dani"},
			want: "Dani <email not set>",
		},
		{
			name: "email only",
			info: Info{Email: "dani@example.com"},
			want: "<name not set> dani@example.com",
		},
		{
			name: "nothing configured",
			info: Info{},
			want: "not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := identityString(tt.info); got != tt.want {
				t.Errorf("identityString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAccountsSummary(t *testing.T) {
	tests := []struct {
		name     string
		accounts []Account
		want     string
	}{
		{
			name:     "no accounts",
			accounts: nil,
			want:     "no accounts detected",
		},
		{
			name: "single account",
			accounts: []Account{
				{Host: "github.com", Username: "Dani-Bhai", Source: "gh"},
			},
			want: "1 account (github.com)",
		},
		{
			name: "accounts on two hosts",
			accounts: []Account{
				{Host: "github.com", Username: "Dani-Bhai", Source: "gh"},
				{Host: "github.com", Username: "Dani-Bhai", Source: "credentials"},
				{Host: "gitlab.com", Username: "dani", Source: "glab"},
			},
			want: "3 accounts (github.com, gitlab.com)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := accountsSummary(tt.accounts); got != tt.want {
				t.Errorf("accountsSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAccountString(t *testing.T) {
	tests := []struct {
		name    string
		account Account
		want    string
	}{
		{
			name:    "active gh account",
			account: Account{Host: "github.com", Username: "Dani-Bhai", Source: "gh", Active: true},
			want:    "github.com: Dani-Bhai (gh CLI, active)",
		},
		{
			name:    "credential store entry",
			account: Account{Host: "gitlab.com", Username: "dani", Source: "credentials"},
			want:    "gitlab.com: dani (credential store)",
		},
		{
			name:    "missing username",
			account: Account{Host: "github.com", Source: "credentials"},
			want:    "github.com: (username not stored) (credential store)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := accountString(tt.account); got != tt.want {
				t.Errorf("accountString() = %q, want %q", got, tt.want)
			}
		})
	}
}
