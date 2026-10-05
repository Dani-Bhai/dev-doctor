//go:build linux

package system

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/Dani-Bhai/dev-doctor/internal/model"
)

type SwapCheck struct{}

func (SwapCheck) Name() string {
	return "Swap"
}

type MemCheck struct{}

func (MemCheck) Name() string {
	return "Memory"
}

type DiskCheck struct{}

func (DiskCheck) Name() string {
	return "Disk Space"
}

func (DiskCheck) Run(ctx context.Context) model.Result {
	var stat syscall.Statfs_t

	err := syscall.Statfs("/", &stat)
	if err != nil {
		return model.Result{
			Name:    "Disk Space",
			Group:   "System",
			Status:  model.StatusError,
			Message: err.Error(),
		}
	}

	free := stat.Bavail * uint64(stat.Bsize)
	total := stat.Blocks * uint64(stat.Bsize)

	freeGb := float64(free) / 1024 / 1024 / 1024
	totalGb := float64(total) / 1024 / 1024 / 1024

	status := model.StatusOk

	if totalGb > 500 && freeGb < totalGb*0.08 || freeGb < 10 {
		status = model.StatusWarning
	}

	return model.Result{
		Name:    "Disk Space",
		Group:   "System",
		Status:  status,
		Message: fmt.Sprintf("%.1f GB free / %.1f GB", freeGb, totalGb),
	}
}

func (MemCheck) Run(ctx context.Context) model.Result {

	values, err := readMemInfo()
	if err != nil {
		return model.Result{
			Name:    "Memory",
			Group:   "System",
			Status:  model.StatusError,
			Message: err.Error(),
		}
	}

	totalKb := values["MemTotal"]
	availableKb := values["MemAvailable"]

	totalGb := float64(totalKb) / 1024 / 1024
	availableGb := float64(availableKb) / 1024 / 1024

	status := model.StatusOk

	usagePercentage := 100 - (float64(availableKb)/float64(totalKb))*100

	if usagePercentage > 90 {
		status = model.StatusWarning
	}

	if totalGb < 16 {
		return model.Result{
			Name:   "Memory",
			Group:  "System",
			Status: status,
			Message: fmt.Sprintf(
				"%.1f GB available / %.1f GB (32 gigs of ram is preferred for heavy dev workloads)",
				availableGb,
				totalGb,
			),
		}
	}

	return model.Result{
		Name:   "Memory",
		Group:  "System",
		Status: status,
		Message: fmt.Sprintf(
			"%.1f GB available / %.1f GB",
			availableGb,
			totalGb,
		),
	}
}

func (SwapCheck) Run(ctx context.Context) model.Result {
	values, err := readMemInfo()
	if err != nil {
		return model.Result{
			Name:    "Swap",
			Group:   "System",
			Status:  model.StatusError,
			Message: err.Error(),
		}
	}

	totalKB := values["SwapTotal"]
	freeKB := values["SwapFree"]

	totalMB := float64(totalKB) / 1024
	freeMB := float64(freeKB) / 1024

	status := model.StatusOk

	switch {
	case totalKB == 0:
		status = model.StatusWarning

	case totalMB < 1024:
		status = model.StatusWarning
	}

	return model.Result{
		Name:   "Swap",
		Group:  "System",
		Status: status,
		Message: fmt.Sprintf(
			"%.0f MB free / %.0f MB",
			freeMB,
			totalMB,
		),
	}
}

func readMemInfo() (map[string]uint64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	values := make(map[string]uint64)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		values[key] = value

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return values, nil
}
