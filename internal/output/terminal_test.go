package output

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Dani-Bhai/dev-doctor/internal/model"
)

func capture(t *testing.T, results []model.Result, verbose bool) string {
	t.Helper()

	original := os.Stdout

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = writer

	Print(results, verbose)

	writer.Close()
	os.Stdout = original

	out, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	reader.Close()

	return string(out)
}

func TestPrintCompactUsesSummary(t *testing.T) {
	results := []model.Result{
		{
			Name:    "Git",
			Group:   "Git",
			Status:  model.StatusOk,
			Message: "2.55.0",
			Summary: []string{"  summary line"},
			Details: []string{"  detail line"},
		},
	}

	out := capture(t, results, false)

	if !strings.Contains(out, "summary line") {
		t.Errorf("compact output does not contain the summary: %q", out)
	}

	if strings.Contains(out, "detail line") {
		t.Errorf("compact output contains details: %q", out)
	}
}

func TestPrintVerboseUsesDetails(t *testing.T) {
	results := []model.Result{
		{
			Name:    "Git",
			Group:   "Git",
			Status:  model.StatusOk,
			Message: "2.55.0",
			Summary: []string{"  summary line"},
			Details: []string{"  detail line"},
		},
	}

	out := capture(t, results, true)

	if !strings.Contains(out, "detail line") {
		t.Errorf("verbose output does not contain the details: %q", out)
	}

	if strings.Contains(out, "summary line") {
		t.Errorf("verbose output contains the summary: %q", out)
	}
}

func TestPrintIconsAndGroups(t *testing.T) {
	results := []model.Result{
		{Name: "Git", Group: "Git", Status: model.StatusOk, Message: "2.55.0"},
		{Name: "Docker", Group: "Docker", Status: model.StatusWarning, Message: "daemon not running"},
		{Name: "Go", Group: "Runtime", Status: model.StatusError, Message: "not found"},
	}

	out := capture(t, results, false)

	for _, want := range []string{"Git\n", "Docker\n", "Runtime\n", "✓ Git", "⚠ Docker", "✗ Go"} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not contain %q:\n%s", want, out)
		}
	}
}
