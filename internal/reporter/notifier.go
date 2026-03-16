package reporter

import "context"

// Notifier sends aggregated CSP violation reports to an external service.
type Notifier interface {
	Send(ctx context.Context, rpt *report) error
	Name() string
}
