package reporter

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/tidwall/gjson"

	"github.com/meysam81/vigil/internal/constants"
)

// report holds aggregated CSP violation data for a time window.
type report struct {
	Total      int
	Since      time.Time
	Until      time.Time
	Directives map[string]int
	Origins    map[string]int
	Pages      map[string]int
	Browsers   map[string]int
}

// rankedEntry is a key-count pair used for sorting.
type rankedEntry struct {
	Key   string
	Count int
}

// topN returns the top n entries from a count map, sorted descending.
func topN(m map[string]int, n int) []rankedEntry {
	entries := make([]rankedEntry, 0, len(m))
	for k, v := range m {
		entries = append(entries, rankedEntry{Key: k, Count: v})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})
	if len(entries) > n {
		entries = entries[:n]
	}
	return entries
}

const mgetBatchSize = 200

func (r *Reporter) aggregate(ctx context.Context, since, now time.Time) (*report, error) {
	rpt := &report{
		Since:      since,
		Until:      now,
		Directives: make(map[string]int),
		Origins:    make(map[string]int),
		Pages:      make(map[string]int),
		Browsers:   make(map[string]int),
	}

	keys, err := r.redis.ZRangeArgs(ctx, goredis.ZRangeArgs{
		Key:     constants.TimelineKey,
		Start:   fmt.Sprintf("%d", since.UnixNano()),
		Stop:    fmt.Sprintf("%d", now.UnixNano()),
		ByScore: true,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("querying timeline index: %w", err)
	}

	// Batch MGET in chunks to avoid oversized commands
	for i := 0; i < len(keys); i += mgetBatchSize {
		end := min(i+mgetBatchSize, len(keys))
		chunk := keys[i:end]

		vals, err := r.redis.MGet(ctx, chunk...).Result()
		if err != nil {
			return nil, fmt.Errorf("fetching report batch: %w", err)
		}

		for _, val := range vals {
			s, ok := val.(string)
			if !ok || s == "" {
				continue
			}
			parseReport(s, rpt)
		}
	}

	return rpt, nil
}

func parseReport(raw string, rpt *report) {
	rpt.Total++

	// Detect format: legacy has "csp-report" top-level key
	if gjson.Get(raw, "csp-report").Exists() {
		parseLegacy(raw, rpt)
	} else {
		parseModern(raw, rpt)
	}
}

func parseLegacy(raw string, rpt *report) {
	if d := gjson.Get(raw, "csp-report.effective-directive").String(); d != "" {
		rpt.Directives[d]++
	}
	if b := gjson.Get(raw, "csp-report.blocked-uri").String(); b != "" {
		rpt.Origins[extractOrigin(b)]++
	}
	if p := gjson.Get(raw, "csp-report.document-uri").String(); p != "" {
		rpt.Pages[extractPath(p)]++
	}
	// Legacy format has no user_agent field
	rpt.Browsers["Unknown"]++
}

func parseModern(raw string, rpt *report) {
	if d := gjson.Get(raw, "body.effectiveDirective").String(); d != "" {
		rpt.Directives[d]++
	}
	if b := gjson.Get(raw, "body.blockedURL").String(); b != "" {
		rpt.Origins[extractOrigin(b)]++
	}
	if p := gjson.Get(raw, "body.documentURL").String(); p != "" {
		rpt.Pages[extractPath(p)]++
	}
	if ua := gjson.Get(raw, "user_agent").String(); ua != "" {
		rpt.Browsers[browserFamily(ua)]++
	} else {
		rpt.Browsers["Unknown"]++
	}
}

func extractOrigin(rawURL string) string {
	// Handle data: and blob: URI schemes
	if strings.HasPrefix(rawURL, "data:") {
		return "data:"
	}
	if strings.HasPrefix(rawURL, "blob:") {
		return "blob:"
	}
	// Handle special CSP values like "inline", "eval", "self"
	if !strings.Contains(rawURL, "://") {
		return rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Hostname()
}

func extractPath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	p := u.Path
	if p == "" {
		p = "/"
	}
	return p
}

// browserFamily extracts a coarse browser name from a User-Agent string.
// This is an intentional approximation: UA strings are unreliable and
// spoofable, so a rough bucket (Chrome/Firefox/Safari/Edge/Opera/Other)
// is sufficient for aggregate reporting.
func browserFamily(ua string) string {
	switch {
	case strings.Contains(ua, "Edg"):
		return "Edge"
	case strings.Contains(ua, "OPR"):
		return "Opera"
	case strings.Contains(ua, "Firefox"):
		return "Firefox"
	case strings.Contains(ua, "Chrome"):
		return "Chrome"
	case strings.Contains(ua, "Safari"):
		return "Safari"
	default:
		return "Other"
	}
}
