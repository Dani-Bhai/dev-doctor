//go:build linux

package main

import (
	"github.com/Dani-Bhai/dev-doctor/internal/checks/docker"
	"github.com/Dani-Bhai/dev-doctor/internal/checks/git"
	"github.com/Dani-Bhai/dev-doctor/internal/checks/runtime"
	"github.com/Dani-Bhai/dev-doctor/internal/checks/system"
	"github.com/Dani-Bhai/dev-doctor/internal/runner"
)

type command struct {
	name  string
	check runner.Check
}

func commands() []command {
	return []command{
		{"node", runtime.NodeCheck{}},
		{"java", runtime.JavaCheck{}},
		{"python", runtime.PythonCheck{}},
		{"go", runtime.GoCheck{}},
		{"git", git.GitCheck{}},
		{"docker", docker.DockerCheck{}},
		{"disk", system.DiskCheck{}},
		{"memory", system.MemCheck{}},
		{"swap", system.SwapCheck{}},
	}
}

func allChecks() []runner.Check {
	cmds := commands()

	checks := make([]runner.Check, 0, len(cmds))

	for _, cmd := range cmds {
		checks = append(checks, cmd.check)
	}

	return checks
}

func findCheck(name string) (runner.Check, bool) {
	for _, cmd := range commands() {
		if cmd.name == name {
			return cmd.check, true
		}
	}

	return nil, false
}
