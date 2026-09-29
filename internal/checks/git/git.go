package git

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Dani-Bhai/dev-doctor/internal/checks/system"
	"github.com/Dani-Bhai/dev-doctor/internal/model"
)

// Account is a git hosting account devdoctor could find on this machine.
// Tokens are never stored or printed, only host and username.
type Account struct {
	Host     string
	Username string
	Source   string
	Active   bool
}

type Info struct {
	Version  string
	Path     string
	Name     string
	Email    string
	Helpers  []string
	Accounts []Account
}

var (
	gitVersionPattern  = regexp.MustCompile(`git version (\S+)`)
	ghAccountPattern   = regexp.MustCompile(`Logged in to (\S+) account (\S+)`)
	glabAccountPattern = regexp.MustCompile(`Logged in to (\S+) as (\S+)`)
)

func parseGitVersion(output string) string {
	matches := gitVersionPattern.FindStringSubmatch(output)

	if len(matches) != 2 {
		return ""
	}

	return matches[1]
}

func getGitVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "--version")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	if version := parseGitVersion(string(out)); version != "" {
		return version, nil
	}

	return strings.TrimSpace(string(out)), nil
}

func getGlobalConfig(ctx context.Context, key string) string {
	cmd := exec.CommandContext(ctx, "git", "config", "--global", "--get", key)

	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

func getCredentialHelpers(ctx context.Context) []string {
	cmd := exec.CommandContext(ctx, "git", "config", "--global", "--get-all", "credential.helper")

	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var helpers []string

	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line == "" {
			continue
		}

		helpers = append(helpers, line)
	}

	return helpers
}

func commandOutput(ctx context.Context, name string, args ...string) string {
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil && len(out) == 0 {
		return ""
	}

	return string(out)
}

// parseGhAccounts reads the output of "gh auth status". Every logged in
// account is followed by an "Active account" line for the host it belongs to.
func parseGhAccounts(output string) []Account {
	var accounts []Account

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)

		if matches := ghAccountPattern.FindStringSubmatch(line); len(matches) == 3 {
			accounts = append(accounts, Account{
				Host:     matches[1],
				Username: matches[2],
				Source:   "gh",
			})

			continue
		}

		if strings.HasPrefix(line, "- Active account: true") && len(accounts) > 0 {
			accounts[len(accounts)-1].Active = true
		}
	}

	return accounts
}

// parseGlabAccounts reads the output of "glab auth status".
func parseGlabAccounts(output string) []Account {
	var accounts []Account

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)

		if matches := glabAccountPattern.FindStringSubmatch(line); len(matches) == 3 {
			accounts = append(accounts, Account{
				Host:     matches[1],
				Username: matches[2],
				Source:   "glab",
			})
		}
	}

	return accounts
}

func credentialStorePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".git-credentials")
}

// readCredentialStore extracts host and username from the plaintext
// "store" helper file (~/.git-credentials). Passwords are discarded.
func readCredentialStore(path string) []Account {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var accounts []Account

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		parsed, err := url.Parse(line)
		if err != nil || parsed.Hostname() == "" {
			continue
		}

		username := ""

		if parsed.User != nil {
			username = parsed.User.Username()
		}

		accounts = append(accounts, Account{
			Host:     parsed.Hostname(),
			Username: username,
			Source:   "credentials",
		})
	}

	return accounts
}

func dedupeAccounts(accounts []Account) []Account {
	seen := make(map[string]struct{}, len(accounts))
	unique := make([]Account, 0, len(accounts))

	for _, account := range accounts {
		key := account.Host + "\x00" + account.Username + "\x00" + account.Source

		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		unique = append(unique, account)
	}

	return unique
}

func detectAccounts(ctx context.Context) []Account {
	var accounts []Account

	if system.CommandExists("gh") {
		accounts = append(accounts, parseGhAccounts(commandOutput(ctx, "gh", "auth", "status"))...)
	}

	if system.CommandExists("glab") {
		accounts = append(accounts, parseGlabAccounts(commandOutput(ctx, "glab", "auth", "status"))...)
	}

	accounts = append(accounts, readCredentialStore(credentialStorePath())...)

	return dedupeAccounts(accounts)
}

// InspectGit collects the git version, the global identity and every hosting
// account it can find through the gh/glab CLIs or the credential store.
func InspectGit(ctx context.Context) Info {
	info := Info{}

	if version, err := getGitVersion(ctx); err == nil {
		info.Version = version
	}

	if path, err := exec.LookPath("git"); err == nil {
		info.Path = path
	}

	if info.Version == "" {
		return info
	}

	info.Name = getGlobalConfig(ctx, "user.name")
	info.Email = getGlobalConfig(ctx, "user.email")
	info.Helpers = getCredentialHelpers(ctx)
	info.Accounts = detectAccounts(ctx)

	return info
}

// GitCheck reports the git version, the configured commit identity and the
// accounts that are logged in on this machine.
type GitCheck struct{}

func (GitCheck) Name() string {
	return "Git"
}

func (GitCheck) Run(ctx context.Context) model.Result {
	info := InspectGit(ctx)

	result := model.Result{
		Name:  "Git",
		Group: "Git",
	}

	if info.Version == "" {
		result.Status = model.StatusError
		result.Message = "not found"

		return result
	}

	result.Status = model.StatusOk
	result.Message = info.Version
	result.Summary = []string{"  " + identityString(info) + " · " + accountsSummary(info.Accounts)}

	if info.Path != "" {
		result.Details = append(result.Details, "  Path: "+info.Path)
	}

	result.Details = append(result.Details, "  Identity: "+identityString(info))

	if len(info.Helpers) > 0 {
		result.Details = append(result.Details, "  Credential helper: "+strings.Join(info.Helpers, ", "))
	}

	result.Details = append(result.Details, "", "Logged in accounts:")

	if len(info.Accounts) == 0 {
		result.Details = append(result.Details, "  none detected")
	} else {
		for _, account := range info.Accounts {
			result.Details = append(result.Details, "  - "+accountString(account))
		}
	}

	if info.Name == "" || info.Email == "" {
		result.Status = model.StatusWarning
	}

	return result
}

func identityString(info Info) string {
	switch {
	case info.Name != "" && info.Email != "":
		return fmt.Sprintf("%s <%s>", info.Name, info.Email)

	case info.Name != "":
		return info.Name + " <email not set>"

	case info.Email != "":
		return "<name not set> " + info.Email

	default:
		return "not configured"
	}
}

// accountsSummary condenses the account list into one line, e.g.
// "2 accounts (github.com, gitlab.com)".
func accountsSummary(accounts []Account) string {
	if len(accounts) == 0 {
		return "no accounts detected"
	}

	hosts := make([]string, 0, len(accounts))
	seen := make(map[string]struct{}, len(accounts))

	for _, account := range accounts {
		if _, ok := seen[account.Host]; ok {
			continue
		}

		seen[account.Host] = struct{}{}
		hosts = append(hosts, account.Host)
	}

	word := "accounts"
	if len(accounts) == 1 {
		word = "account"
	}

	return fmt.Sprintf("%d %s (%s)", len(accounts), word, strings.Join(hosts, ", "))
}

func accountString(account Account) string {
	label := account.Host

	if account.Username != "" {
		label += ": " + account.Username
	} else {
		label += ": (username not stored)"
	}

	source := "credential store"

	switch account.Source {
	case "gh":
		source = "gh CLI"

	case "glab":
		source = "glab CLI"
	}

	if account.Active {
		source += ", active"
	}

	return fmt.Sprintf("%s (%s)", label, source)
}
