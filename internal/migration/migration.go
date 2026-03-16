package migration

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/meysam81/vigil/internal/constants"
	"github.com/meysam81/vigil/internal/logger"
)

const (
	stateKey  = "vigil:migration:timeline_backfill"
	scanBatch = 500
)

// Run backfills vigil:timeline from existing csp:* keys.
// Idempotent: skips if the migration state key already exists in Redis.
func Run(ctx context.Context, redis *goredis.Client, log *logger.Logger) error {
	exists, err := redis.Exists(ctx, stateKey).Result()
	if err != nil {
		return fmt.Errorf("checking migration state: %w", err)
	}
	if exists > 0 {
		log.Debug().Str("state_key", stateKey).Msg("timeline backfill already completed, skipping migration")
		return nil
	}

	log.Info().Msg("starting timeline backfill migration")

	var cursor uint64
	var total int

	for {
		keys, nextCursor, err := redis.Scan(ctx, cursor, "csp:*", scanBatch).Result()
		if err != nil {
			return fmt.Errorf("scanning keys: %w", err)
		}

		if len(keys) > 0 {
			pipe := redis.Pipeline()
			added := 0
			for _, key := range keys {
				nanos, err := parseNanosFromKey(key)
				if err != nil {
					log.Warn().Str("key", key).Err(err).Msg("skipping key with invalid format")
					continue
				}
				pipe.ZAdd(ctx, constants.TimelineKey, goredis.Z{
					Score:  float64(nanos),
					Member: key,
				})
				added++
			}
			if added > 0 {
				if _, err := pipe.Exec(ctx); err != nil {
					log.Error().Err(err).Int("batch_size", added).Msg("failed adding timeline entries")
				} else {
					total += added
				}
			}
		}

		if total > 0 && total%10000 < scanBatch {
			log.Info().Int("migrated", total).Msg("migration progress")
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	if err := redis.Set(ctx, stateKey, time.Now().Unix(), 0).Err(); err != nil {
		log.Error().Err(err).Msg("failed saving migration state (migration data is already in Redis)")
	}

	log.Info().Int("total", total).Msg("timeline backfill migration complete")
	return nil
}

// parseNanosFromKey extracts the UnixNano timestamp from a key with format csp:<nanos>:<rand>.
func parseNanosFromKey(key string) (int64, error) {
	parts := strings.SplitN(key, ":", 3)
	if len(parts) != 3 || parts[0] != "csp" {
		return 0, fmt.Errorf("invalid key format: %s", key)
	}
	nanos, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing nanos from key %s: %w", key, err)
	}
	return nanos, nil
}
