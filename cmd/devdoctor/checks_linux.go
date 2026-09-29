//go:build linux

package main

import (
	"github.com/Dani-Bhai/dev-doctor/internal/checks/runtime"
	"github.com/Dani-Bhai/dev-doctor/internal/checks/system"
	"github.com/Dani-Bhai/dev-doctor/internal/runner"
)

func checks() []runner.Check {
	return []runner.Check{
		runtime.NodeCheck{},
		runtime.JavaCheck{},
		runtime.PythonCheck{},
		runtime.GoCheck{},
		system.DiskCheck{},
		system.MemCheck{},
		system.SwapCheck{},
	}
}
