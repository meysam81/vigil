package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/httperr"
	"github.com/meysam81/vigil/internal/logger"
	"github.com/meysam81/x/ratelimit"
)

func RateLimitMiddleware(redisClient *goredis.Client, cfg *config.RateLimitConfig, log *logger.Logger) func(http.Handler) http.Handler {
	rl := ratelimit.RateLimit{
		Redis:       redisClient,
		MaxRequests: cfg.MaxRPS,
		RefillRate:  cfg.RefillRate,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				clientIP = strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
			} else if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
				clientIP = realIP
			}

			rate := rl.TokenBucket(r.Context(), clientIP)

			w.Header().Set("X-Ratelimit-Total", fmt.Sprintf("%d", rate.Total))
			w.Header().Set("X-Ratelimit-Remaining", fmt.Sprintf("%d", rate.Remaining))

			if !rate.Allowed {
				log.Warn().Str("client_ip", clientIP).Int64("remaining", rate.Remaining).Msg("rate limit exceeded")
				retryAfter := int(time.Until(rate.ResetAt()).Seconds())
				if retryAfter < 1 {
					retryAfter = 1
				}
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				httperr.TooManyRequests(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
