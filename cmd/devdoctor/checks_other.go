//go:build !linux

package main

import "github.com/Dani-Bhai/dev-doctor/internal/runner"

func allChecks() []runner.Check {
	return nil
}

func findCheck(name string) (runner.Check, bool) {
	return nil, false
}
