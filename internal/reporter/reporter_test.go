package reporter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/goccy/go-json"
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

// mockNotifier records Send calls for testing.
type mockNotifier struct {
	name    string
	sendErr error
	calls   int
}

func (m *mockNotifier) Send(_ context.Context, _ *report) error {
	m.calls++
	return m.sendErr
}

func (m *mockNotifier) Name() string { return m.name }

func TestFilterPending_NoPending(t *testing.T) {
	slack := &mockNotifier{name: "slack"}
	discord := &mockNotifier{name: "discord"}
	all := []Notifier{slack, discord}

	targets := filterPending(all, nil)
	if len(targets) != 2 {
		t.Fatalf("want 2, got %d", len(targets))
	}
}

func TestFilterPending_OnlyDiscord(t *testing.T) {
	slack := &mockNotifier{name: "slack"}
	discord := &mockNotifier{name: "discord"}
	all := []Notifier{slack, discord}

	targets := filterPending(all, []string{"discord"})
	if len(targets) != 1 {
		t.Fatalf("want 1, got %d", len(targets))
	}
	if targets[0].Name() != "discord" {
		t.Errorf("want discord, got %s", targets[0].Name())
	}
}

func TestFilterPending_UnknownName(t *testing.T) {
	slack := &mockNotifier{name: "slack"}
	all := []Notifier{slack}

	targets := filterPending(all, []string{"telegram"})
	if len(targets) != 0 {
		t.Fatalf("want 0 targets for unknown pending name, got %d", len(targets))
	}
}

func TestFilterPending_AllPending(t *testing.T) {
	slack := &mockNotifier{name: "slack"}
	discord := &mockNotifier{name: "discord"}
	all := []Notifier{slack, discord}

	targets := filterPending(all, []string{"slack", "discord"})
	if len(targets) != 2 {
		t.Fatalf("want 2, got %d", len(targets))
	}
}

func TestMockNotifier_SatisfiesInterface(t *testing.T) {
	var n Notifier = &mockNotifier{name: "test"}
	if n.Name() != "test" {
		t.Errorf("Name: want test, got %s", n.Name())
	}
	if err := n.Send(context.Background(), &report{}); err != nil {
		t.Errorf("Send: unexpected error: %v", err)
	}
}

func TestMockNotifier_ErrorPropagation(t *testing.T) {
	errFail := fmt.Errorf("boom")
	n := &mockNotifier{name: "failing", sendErr: errFail}

	err := n.Send(context.Background(), &report{})
	if err != errFail {
		t.Errorf("Send: want errFail, got %v", err)
	}
	if n.calls != 1 {
		t.Errorf("calls: want 1, got %d", n.calls)
	}
}

func TestState_PendingField_MarshalRoundTrip(t *testing.T) {
	st := &state{
		LastSuccessAt: 1000,
		LastAttemptAt: 2000,
		Status:        statusFailed,
		Error:         "failed notifiers: discord",
		Count:         5,
		Pending:       []string{"discord"},
	}

	data, err := json.Marshal(st)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got state
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got.Pending) != 1 || got.Pending[0] != "discord" {
		t.Errorf("Pending: want [discord], got %v", got.Pending)
	}
	if got.LastSuccessAt != 1000 {
		t.Errorf("LastSuccessAt: want 1000, got %d", got.LastSuccessAt)
	}
}

func TestState_PendingOmittedWhenEmpty(t *testing.T) {
	st := &state{
		LastSuccessAt: 1000,
		Status:        statusSuccess,
		Count:         1,
	}

	data, err := json.Marshal(st)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Pending should not appear in JSON
	raw := string(data)
	if contains(raw, "pending") {
		t.Errorf("empty Pending should be omitted from JSON, got: %s", raw)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && stringContains(s, substr)
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
