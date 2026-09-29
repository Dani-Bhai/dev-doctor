package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/Dani-Bhai/dev-doctor/internal/model"
)

const (
	infoFormat  = "{{.ServerVersion}}|{{.OperatingSystem}}|{{.Driver}}|{{.ContainersRunning}}|{{.Containers}}|{{.Images}}"
	usageFormat = "{{.Type}}|{{.TotalCount}}|{{.Size}}|{{.Reclaimable}}"
	psFormat    = "{{.Names}}|{{.Image}}|{{.Status}}"

	maxContainers = 10
)

type Container struct {
	Name   string
	Image  string
	Status string
}

type DiskUsage struct {
	Type        string
	Count       string
	Size        string
	Reclaimable string
}

type daemonInfo struct {
	serverVersion     string
	operatingSystem   string
	driver            string
	containersRunning int
	containers        int
	images            int
}

type Info struct {
	Version       string
	Path          string
	Context       string
	DaemonRunning bool
	DaemonError   string
	ServerVersion string
	ServerOS      string
	Driver        string
	RunningCount  int
	TotalCount    int
	ImageCount    int
	Usage         []DiskUsage
	Containers    []Container
}

var dockerVersionPattern = regexp.MustCompile(`Docker version (\S+),`)

func parseClientVersion(output string) string {
	matches := dockerVersionPattern.FindStringSubmatch(output)

	if len(matches) != 2 {
		return ""
	}

	return matches[1]
}

func getClientVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "--version")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	if version := parseClientVersion(string(out)); version != "" {
		return version, nil
	}

	return strings.TrimSpace(string(out)), nil
}

func getDockerContext(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "context", "show")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func getDaemonInfo(ctx context.Context) (daemonInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "info", "--format", infoFormat)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return daemonInfo{}, errors.New(daemonErrorMessage(string(out), stderr.String(), err))
	}

	info, err := parseDaemonInfo(string(out))
	if err != nil {
		return daemonInfo{}, err
	}

	if info.serverVersion == "" {
		return daemonInfo{}, errors.New("could not determine docker server version")
	}

	return info, nil
}

// daemonErrorMessage prefers stderr because "docker info" writes a partially
// rendered format string to stdout before failing when the daemon is down.
func daemonErrorMessage(stdout, stderr string, fallback error) string {
	if message := firstNonEmptyLine(stderr); message != "" {
		return message
	}

	if message := firstNonEmptyLine(stdout); message != "" {
		return message
	}

	return fallback.Error()
}

func parseDaemonInfo(output string) (daemonInfo, error) {
	fields := strings.Split(strings.TrimSpace(output), "|")

	if len(fields) != 6 {
		return daemonInfo{}, fmt.Errorf("unexpected docker info output: %q", strings.TrimSpace(output))
	}

	return daemonInfo{
		serverVersion:     fields[0],
		operatingSystem:   fields[1],
		driver:            fields[2],
		containersRunning: atoi(fields[3]),
		containers:        atoi(fields[4]),
		images:            atoi(fields[5]),
	}, nil
}

func getUsage(ctx context.Context) []DiskUsage {
	cmd := exec.CommandContext(ctx, "docker", "system", "df", "--format", usageFormat)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	return parseUsage(string(out))
}

func parseUsage(output string) []DiskUsage {
	var usage []DiskUsage

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Split(strings.TrimSpace(line), "|")

		if len(fields) != 4 {
			continue
		}

		usage = append(usage, DiskUsage{
			Type:        fields[0],
			Count:       fields[1],
			Size:        fields[2],
			Reclaimable: fields[3],
		})
	}

	return usage
}

func getContainers(ctx context.Context) []Container {
	cmd := exec.CommandContext(ctx, "docker", "ps", "--format", psFormat)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	return parseContainers(string(out))
}

func parseContainers(output string) []Container {
	var containers []Container

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Split(strings.TrimSpace(line), "|")

		if len(fields) != 3 {
			continue
		}

		containers = append(containers, Container{
			Name:   fields[0],
			Image:  fields[1],
			Status: fields[2],
		})
	}

	return containers
}

// InspectDocker collects the client version, daemon status, context, disk
// usage and the running containers. Every part degrades gracefully when the
// daemon is unreachable.
func InspectDocker(ctx context.Context) Info {
	info := Info{}

	if version, err := getClientVersion(ctx); err == nil {
		info.Version = version
	}

	if path, err := exec.LookPath("docker"); err == nil {
		info.Path = path
	}

	if info.Version == "" {
		return info
	}

	if dockerContext, err := getDockerContext(ctx); err == nil {
		info.Context = dockerContext
	}

	server, err := getDaemonInfo(ctx)
	if err != nil {
		info.DaemonError = err.Error()

		return info
	}

	info.DaemonRunning = true
	info.ServerVersion = server.serverVersion
	info.ServerOS = server.operatingSystem
	info.Driver = server.driver
	info.RunningCount = server.containersRunning
	info.TotalCount = server.containers
	info.ImageCount = server.images
	info.Usage = getUsage(ctx)
	info.Containers = getContainers(ctx)

	return info
}

// DockerCheck reports the docker client version, the daemon status, the
// active context and the current resource usage.
type DockerCheck struct{}

func (DockerCheck) Name() string {
	return "Docker"
}

func (DockerCheck) Run(ctx context.Context) model.Result {
	info := InspectDocker(ctx)

	result := model.Result{
		Name:  "Docker",
		Group: "Docker",
	}

	if info.Version == "" {
		result.Status = model.StatusError
		result.Message = "not found"

		return result
	}

	result.Message = info.Version

	if summary := dockerSummary(info); summary != "" {
		result.Summary = []string{summary}
	}

	if info.Path != "" {
		result.Details = append(result.Details, "  Path: "+info.Path)
	}

	if info.Context != "" {
		result.Details = append(result.Details, "  Context: "+info.Context)
	}

	if !info.DaemonRunning {
		result.Status = model.StatusWarning
		result.Message = info.Version + " (daemon not running)"

		if info.DaemonError != "" {
			result.Details = append(result.Details, "  Daemon: "+info.DaemonError)
		}

		return result
	}

	result.Status = model.StatusOk

	server := info.ServerVersion

	if info.ServerOS != "" {
		server += " (" + info.ServerOS

		if info.Driver != "" {
			server += ", " + info.Driver
		}

		server += ")"
	}

	result.Details = append(result.Details, "  Server: "+server)
	result.Details = append(result.Details, fmt.Sprintf("  Containers: %d running / %d total", info.RunningCount, info.TotalCount))
	result.Details = append(result.Details, fmt.Sprintf("  Images: %d", info.ImageCount))

	if len(info.Usage) > 0 {
		result.Details = append(result.Details, "", "Disk usage:")

		for _, entry := range info.Usage {
			line := fmt.Sprintf("  - %s: %s, %s", entry.Type, entry.Count, entry.Size)

			if entry.Reclaimable != "" {
				line += ", reclaimable " + entry.Reclaimable
			}

			result.Details = append(result.Details, line)
		}
	}

	if len(info.Containers) > 0 {
		result.Details = append(result.Details, "", "Running containers:")

		for i, container := range info.Containers {
			if i == maxContainers {
				result.Details = append(result.Details, fmt.Sprintf("  ... and %d more", len(info.Containers)-maxContainers))

				break
			}

			result.Details = append(result.Details, fmt.Sprintf("  - %s (%s) %s", container.Name, container.Image, container.Status))
		}
	}

	return result
}

// dockerSummary condenses the daemon state into one line for compact output.
// It returns an empty string when the daemon is down, because the message
// already says so.
func dockerSummary(info Info) string {
	if !info.DaemonRunning {
		return ""
	}

	return fmt.Sprintf(
		"  daemon running · %d running / %d total containers · %d images",
		info.RunningCount,
		info.TotalCount,
		info.ImageCount,
	)
}

func atoi(value string) int {
	number, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}

	return number
}

func firstNonEmptyLine(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}

	return ""
}
