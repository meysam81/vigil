package reporter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	goredis "github.com/redis/go-redis/v9"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/logger"
)

const stateKey = "vigil:reporter:state"

const (
	statusSuccess = "success"
	statusFailed  = "failed"
)

type state struct {
	LastSuccessAt int64  `json:"last_success_at"`
	LastAttemptAt int64  `json:"last_attempt_at"`
	Status        string `json:"status"`
	Error         string `json:"error,omitempty"`
	Count         int    `json:"count"`
}

// Reporter runs periodic CSP violation aggregate reports and sends them via Notifiers.
type Reporter struct {
	redis     *goredis.Client
	log       *logger.Logger
	cfg       *config.SlackConfig
	notifiers []Notifier
}

func New(redis *goredis.Client, log *logger.Logger, cfg *config.SlackConfig, notifiers ...Notifier) *Reporter {
	return &Reporter{redis: redis, log: log, cfg: cfg, notifiers: notifiers}
}

// Start runs the reporting loop until ctx is cancelled. It checks for overdue
// reports on startup (crash recovery) and then ticks at the configured interval.
func (r *Reporter) Start(ctx context.Context) error {
	st, err := r.loadState(ctx)
	if err != nil {
		r.log.Warn().Err(err).Msg("failed loading reporter state, treating as first run")
		st = &state{}
	}

	if r.isOverdue(st) {
		r.log.Info().Msg("reporter overdue, sending immediate report")
		r.runCycle(ctx, st)
	}

	ticker := time.NewTicker(r.cfg.ReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.log.Info().Msg("reporter shutting down")
			return nil
		case <-ticker.C:
			st, err = r.loadState(ctx)
			if err != nil {
				r.log.Error().Err(err).Msg("failed loading reporter state")
				continue
			}
			r.runCycle(ctx, st)
		}
	}
}

func (r *Reporter) isOverdue(st *state) bool {
	if st.LastSuccessAt == 0 {
		return true // first run
	}
	if st.Status == statusFailed {
		return true // last attempt failed, retry
	}
	lastSuccess := time.Unix(st.LastSuccessAt, 0)
	return time.Since(lastSuccess) > r.cfg.ReportInterval
}

func (r *Reporter) runCycle(ctx context.Context, st *state) {
	now := time.Now()

	since := time.Unix(st.LastSuccessAt, 0)
	if st.LastSuccessAt == 0 {
		since = now.Add(-r.cfg.ReportInterval)
	}

	rpt, err := r.aggregate(ctx, since, now)
	if err != nil {
		r.log.Error().Err(err).Msg("failed aggregating reports")
		if sErr := r.saveState(ctx, &state{
			LastAttemptAt: now.Unix(),
			LastSuccessAt: st.LastSuccessAt,
			Status:        statusFailed,
			Error:         err.Error(),
			Count:         st.Count,
		}); sErr != nil {
			r.log.Error().Err(sErr).Msg("failed saving reporter state")
		}
		return
	}

	if rpt.Total == 0 {
		r.log.Info().Msg("no violations in reporting window, skipping notification")
		if sErr := r.saveState(ctx, &state{
			LastAttemptAt: now.Unix(),
			LastSuccessAt: now.Unix(),
			Status:        statusSuccess,
			Count:         st.Count,
		}); sErr != nil {
			r.log.Error().Err(sErr).Msg("failed saving reporter state")
		}
		return
	}

	var sendErrs []error
	for _, n := range r.notifiers {
		if err := n.Send(ctx, rpt); err != nil {
			r.log.Error().Err(err).Str("notifier", n.Name()).Msg("notifier failed")
			sendErrs = append(sendErrs, fmt.Errorf("%s: %w", n.Name(), err))
		}
	}

	if err := errors.Join(sendErrs...); err != nil {
		if sErr := r.saveState(ctx, &state{
			LastAttemptAt: now.Unix(),
			LastSuccessAt: st.LastSuccessAt,
			Status:        statusFailed,
			Error:         err.Error(),
			Count:         st.Count,
		}); sErr != nil {
			r.log.Error().Err(sErr).Msg("failed saving reporter state")
		}
		return
	}

	r.log.Info().Int("violations", rpt.Total).Msg("report sent successfully")
	if sErr := r.saveState(ctx, &state{
		LastAttemptAt: now.Unix(),
		LastSuccessAt: now.Unix(),
		Status:        statusSuccess,
		Count:         st.Count + 1,
	}); sErr != nil {
		r.log.Error().Err(sErr).Msg("failed saving reporter state")
	}
}

func (r *Reporter) loadState(ctx context.Context) (*state, error) {
	val, err := r.redis.Get(ctx, stateKey).Result()
	if errors.Is(err, goredis.Nil) {
		return &state{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading reporter state: %w", err)
	}

	var st state
	if err := json.Unmarshal([]byte(val), &st); err != nil {
		return nil, fmt.Errorf("unmarshaling reporter state: %w", err)
	}
	return &st, nil
}

func (r *Reporter) saveState(ctx context.Context, st *state) error {
	data, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("marshaling reporter state: %w", err)
	}
	if err := r.redis.Set(ctx, stateKey, data, 0).Err(); err != nil {
		return fmt.Errorf("saving reporter state: %w", err)
	}
	return nil
}
