package reporter

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	goredis "github.com/redis/go-redis/v9"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/constants"
	"github.com/meysam81/vigil/internal/logger"
)

const stateKey = "vigil:reporter:state"

const (
	statusSuccess = "success"
	statusFailed  = "failed"
)

type state struct {
	LastSuccessAt int64    `json:"last_success_at"`
	LastAttemptAt int64    `json:"last_attempt_at"`
	Status        string   `json:"status"`
	Error         string   `json:"error,omitempty"`
	Count         int      `json:"count"`
	Pending       []string `json:"pending,omitempty"`
}

// Reporter runs periodic CSP violation aggregate reports and sends them via Notifiers.
type Reporter struct {
	redis     *goredis.Client
	log       *logger.Logger
	cfg       *config.ReporterConfig
	keyTTL    time.Duration
	notifiers []Notifier
	nowFunc   func() time.Time
}

func New(redis *goredis.Client, log *logger.Logger, cfg *config.ReporterConfig, keyTTL time.Duration, notifiers ...Notifier) *Reporter {
	return &Reporter{redis: redis, log: log, cfg: cfg, keyTTL: keyTTL, notifiers: notifiers, nowFunc: time.Now}
}

func (r *Reporter) now() time.Time { return r.nowFunc() }

// nextFireTime computes the next UTC time at the given hour:minute.
// If today's scheduled time has already passed (or is exactly now), it returns tomorrow's.
func nextFireTime(now time.Time, hour, minute int) time.Time {
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC)
	if t.After(now) {
		return t
	}
	return t.AddDate(0, 0, 1)
}

// Start runs the reporting loop until ctx is cancelled.
// Reports fire daily at the configured UTC hour:minute.
func (r *Reporter) Start(ctx context.Context) error {
	r.log.Info().
		Int("schedule_hour", r.cfg.ScheduleHour).
		Int("schedule_minute", r.cfg.ScheduleMin).
		Msg("reporter started")

	for {
		now := r.now()
		next := nextFireTime(now, r.cfg.ScheduleHour, r.cfg.ScheduleMin)
		delay := next.Sub(now)

		r.log.Info().Time("next_fire", next).Dur("delay", delay).Msg("waiting for next report")

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			r.log.Info().Msg("reporter shutting down")
			return nil
		case <-timer.C:
			st, err := r.loadState(ctx)
			if err != nil {
				r.log.Error().Err(err).Msg("failed loading reporter state")
				continue
			}
			r.runCycle(ctx, st)
		}
	}
}

func (r *Reporter) runCycle(ctx context.Context, st *state) {
	now := r.now()

	since := time.Unix(st.LastSuccessAt, 0)
	if st.LastSuccessAt == 0 {
		since = now.Add(-24 * time.Hour)
	}

	r.log.Debug().Time("since", since).Time("until", now).Msg("reporter cycle starting")

	rpt, err := r.aggregate(ctx, since, now)
	if err != nil {
		r.log.Error().Err(err).Msg("failed aggregating reports")
		if sErr := r.saveState(ctx, &state{
			LastAttemptAt: now.Unix(),
			LastSuccessAt: st.LastSuccessAt,
			Status:        statusFailed,
			Error:         err.Error(),
			Count:         st.Count,
			Pending:       st.Pending,
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

	// Determine which notifiers to send to.
	// If previous cycle had partial failures, only retry those.
	targets := filterPending(r.notifiers, st.Pending)

	var failed []string
	for _, n := range targets {
		if err := n.Send(ctx, rpt); err != nil {
			r.log.Error().Err(err).Str("notifier", n.Name()).Msg("notifier failed")
			failed = append(failed, n.Name())
		}
	}

	if len(failed) > 0 {
		if sErr := r.saveState(ctx, &state{
			LastAttemptAt: now.Unix(),
			LastSuccessAt: st.LastSuccessAt,
			Status:        statusFailed,
			Error:         fmt.Sprintf("failed notifiers: %s", strings.Join(failed, ", ")),
			Count:         st.Count,
			Pending:       failed,
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

	r.pruneTimeline(ctx, now)
}

// filterPending returns only the notifiers whose names appear in pending.
// If pending is empty, all notifiers are returned.
func filterPending(all []Notifier, pending []string) []Notifier {
	if len(pending) == 0 {
		return all
	}
	set := make(map[string]bool, len(pending))
	for _, name := range pending {
		set[name] = true
	}
	filtered := make([]Notifier, 0, len(pending))
	for _, n := range all {
		if set[n.Name()] {
			filtered = append(filtered, n)
		}
	}
	return filtered
}

// pruneTimeline removes expired entries from the timeline sorted set.
// Report data keys expire via Redis TTL, but their sorted set members
// remain as orphans. This prevents unbounded memory growth.
func (r *Reporter) pruneTimeline(ctx context.Context, now time.Time) {
	cutoff := now.Add(-r.keyTTL).UnixNano()
	removed, err := r.redis.ZRemRangeByScore(ctx, constants.TimelineKey, "-inf", strconv.FormatInt(cutoff, 10)).Result()
	if err != nil {
		r.log.Error().Err(err).Msg("failed pruning timeline")
		return
	}
	if removed > 0 {
		r.log.Info().Int64("removed", removed).Msg("pruned expired timeline entries")
	}
}

func (r *Reporter) loadState(ctx context.Context) (*state, error) {
	val, err := r.redis.Get(ctx, stateKey).Result()
	if err != nil {
		if goredis.Nil == err {
			return &state{}, nil
		}
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
