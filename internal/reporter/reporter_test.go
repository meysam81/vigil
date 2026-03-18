package reporter

import (
	"testing"
	"time"
)

func TestNextFireTime(t *testing.T) {
	tests := []struct {
		name   string
		now    time.Time
		hour   int
		minute int
		want   time.Time
	}{
		{
			"before schedule today",
			time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC),
			10, 0,
			time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			"after schedule today rolls to tomorrow",
			time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			10, 0,
			time.Date(2025, 1, 16, 10, 0, 0, 0, time.UTC),
		},
		{
			"exactly at schedule rolls to tomorrow",
			time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
			10, 0,
			time.Date(2025, 1, 16, 10, 0, 0, 0, time.UTC),
		},
		{
			"one second before schedule",
			time.Date(2025, 1, 15, 9, 59, 59, 0, time.UTC),
			10, 0,
			time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			"midnight schedule",
			time.Date(2025, 1, 15, 23, 30, 0, 0, time.UTC),
			0, 0,
			time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC),
		},
		{
			"month boundary rollover",
			time.Date(2025, 1, 31, 14, 0, 0, 0, time.UTC),
			10, 0,
			time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			"year boundary rollover",
			time.Date(2025, 12, 31, 14, 0, 0, 0, time.UTC),
			10, 0,
			time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			"custom minute before",
			time.Date(2025, 6, 1, 10, 29, 0, 0, time.UTC),
			10, 30,
			time.Date(2025, 6, 1, 10, 30, 0, 0, time.UTC),
		},
		{
			"custom minute just past",
			time.Date(2025, 6, 1, 10, 31, 0, 0, time.UTC),
			10, 30,
			time.Date(2025, 6, 2, 10, 30, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextFireTime(tt.now, tt.hour, tt.minute)
			if !got.Equal(tt.want) {
				t.Errorf("nextFireTime(%v, %d, %d) = %v, want %v", tt.now, tt.hour, tt.minute, got, tt.want)
			}
		})
	}
}
