package reporter

import (
	"testing"
	"time"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/logger"
)

func TestIsOverdue(t *testing.T) {
	log := logger.NewLogger("error", true)
	cfg := &config.SlackConfig{
		ReportInterval: 24 * time.Hour,
	}
	r := New(nil, log, cfg, 720*time.Hour)

	tests := []struct {
		name string
		st   *state
		want bool
	}{
		{
			"first run (zero timestamp)",
			&state{LastSuccessAt: 0},
			true,
		},
		{
			"last attempt failed",
			&state{
				LastSuccessAt: time.Now().Add(-1 * time.Hour).Unix(),
				Status:        statusFailed,
			},
			true,
		},
		{
			"stale (> interval)",
			&state{
				LastSuccessAt: time.Now().Add(-25 * time.Hour).Unix(),
				Status:        statusSuccess,
			},
			true,
		},
		{
			"recent success (< interval)",
			&state{
				LastSuccessAt: time.Now().Add(-1 * time.Hour).Unix(),
				Status:        statusSuccess,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.isOverdue(tt.st)
			if got != tt.want {
				t.Errorf("isOverdue(%+v) = %v, want %v", tt.st, got, tt.want)
			}
		})
	}
}
