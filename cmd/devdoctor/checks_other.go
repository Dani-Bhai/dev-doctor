//go:build !linux

package main

import "github.com/Dani-Bhai/dev-doctor/internal/runner"

func checks() []runner.Check {
	return nil
}
