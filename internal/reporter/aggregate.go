package reporter

import (
	"context"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// Report holds aggregated CSP violation data for a time window.
type Report struct {
	Total      int
	Since      time.Time
	Until      time.Time
	Directives map[string]int
	Origins    map[string]int
	Pages      map[string]int
	Browsers   map[string]int
}

// RankedEntry is a key-count pair used for sorting.
type RankedEntry struct {
	Key   string
	Count int
}

// TopN returns the top n entries from a count map, sorted descending.
func TopN(m map[string]int, n int) []RankedEntry {
	entries := make([]RankedEntry, 0, len(m))
	for k, v := range m {
		entries = append(entries, RankedEntry{Key: k, Count: v})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})
	if len(entries) > n {
		entries = entries[:n]
	}
	return entries
}

const scanBatchSize = 200

func (r *Reporter) aggregate(ctx context.Context, since time.Time) (*Report, error) {
	sinceNano := since.UnixNano()
	now := time.Now()

	report := &Report{
		Since:      since,
		Until:      now,
		Directives: make(map[string]int),
		Origins:    make(map[string]int),
		Pages:      make(map[string]int),
		Browsers:   make(map[string]int),
	}

	var cursor uint64
	for {
		keys, nextCursor, err := r.redis.Scan(ctx, cursor, "csp:*", scanBatchSize).Result()
		if err != nil {
			return nil, err
		}

		// Filter keys by timestamp embedded in key name (csp:<nanos>:<rand>)
		var filtered []string
		for _, key := range keys {
			parts := strings.SplitN(key, ":", 3)
			if len(parts) < 2 {
				continue
			}
			nanos, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				continue
			}
			if nanos > sinceNano {
				filtered = append(filtered, key)
			}
		}

		if len(filtered) > 0 {
			vals, err := r.redis.MGet(ctx, filtered...).Result()
			if err != nil {
				return nil, err
			}

			for _, val := range vals {
				s, ok := val.(string)
				if !ok || s == "" {
					continue
				}
				r.parseReport(s, report)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return report, nil
}

func (r *Reporter) parseReport(raw string, report *Report) {
	report.Total++

	// Detect format: legacy has "csp-report" top-level key
	if gjson.Get(raw, "csp-report").Exists() {
		r.parseLegacy(raw, report)
	} else {
		r.parseModern(raw, report)
	}
}

func (r *Reporter) parseLegacy(raw string, report *Report) {
	if d := gjson.Get(raw, "csp-report.effective-directive").String(); d != "" {
		report.Directives[d]++
	}
	if b := gjson.Get(raw, "csp-report.blocked-uri").String(); b != "" {
		report.Origins[extractOrigin(b)]++
	}
	if p := gjson.Get(raw, "csp-report.document-uri").String(); p != "" {
		report.Pages[extractPath(p)]++
	}
	// Legacy format has no user_agent field
	report.Browsers["Unknown"]++
}

func (r *Reporter) parseModern(raw string, report *Report) {
	if d := gjson.Get(raw, "body.effectiveDirective").String(); d != "" {
		report.Directives[d]++
	}
	if b := gjson.Get(raw, "body.blockedURL").String(); b != "" {
		report.Origins[extractOrigin(b)]++
	}
	if p := gjson.Get(raw, "body.documentURL").String(); p != "" {
		report.Pages[extractPath(p)]++
	}
	if ua := gjson.Get(raw, "user_agent").String(); ua != "" {
		report.Browsers[browserFamily(ua)]++
	} else {
		report.Browsers["Unknown"]++
	}
}

func extractOrigin(rawURL string) string {
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
