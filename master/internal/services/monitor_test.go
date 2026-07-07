package services

import (
	"testing"

	"github.com/sneaky-developer/UptimeHub/master/internal/models"
)

func checks(ups ...bool) []models.CheckResult {
	results := make([]models.CheckResult, len(ups))
	for i, up := range ups {
		results[i] = models.CheckResult{IsUp: up}
	}
	return results
}

func TestComputeStatus(t *testing.T) {
	tests := []struct {
		name      string
		checks    []models.CheckResult // newest first
		threshold int
		want      string
	}{
		{"all passing", checks(true, true, true), 3, "up"},
		{"single check passing", checks(true), 3, "up"},
		{"one recent failure", checks(false, true, true), 3, "degraded"},
		{"two consecutive failures below threshold", checks(false, false, true), 3, "degraded"},
		{"threshold consecutive failures", checks(false, false, false), 3, "down"},
		{"recovered after failures", checks(true, false, false), 3, "up"},
		{"old failures do not count", checks(true, false, false, false), 3, "up"},
		{"threshold one goes straight down", checks(false), 1, "down"},
		{"more failures than threshold", checks(false, false, false, false), 3, "down"},
		{"non-consecutive failures stay degraded", checks(false, true, false, false), 3, "degraded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := computeStatus(tt.checks, tt.threshold); got != tt.want {
				t.Errorf("computeStatus(%v, %d) = %q, want %q", tt.checks, tt.threshold, got, tt.want)
			}
		})
	}
}
