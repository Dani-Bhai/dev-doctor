package runtime

import (
	"slices"
	"testing"
)

func TestVersionSummary(t *testing.T) {
	tests := []struct {
		name     string
		managers []ManagerResult
		want     []string
	}{
		{
			name:     "no managers",
			managers: nil,
			want:     nil,
		},
		{
			name: "single version",
			managers: []ManagerResult{
				{Manager: "mise", Versions: []string{"20.0.0"}},
			},
			want: []string{"  1 version found (mise)"},
		},
		{
			name: "multiple managers",
			managers: []ManagerResult{
				{Manager: "mise", Versions: []string{"17", "21"}},
				{Manager: "IntelliJ", Versions: []string{"26"}},
			},
			want: []string{"  3 versions found (mise, IntelliJ)"},
		},
		{
			name: "manager without versions",
			managers: []ManagerResult{
				{Manager: "uv"},
			},
			want: nil,
		},
		{
			name: "manager without versions does not hide the others",
			managers: []ManagerResult{
				{Manager: "mise", Versions: []string{"3.11.15"}},
				{Manager: "uv"},
			},
			want: []string{"  1 version found (mise)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := versionSummary(tt.managers)

			if !slices.Equal(got, tt.want) {
				t.Errorf("versionSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}
