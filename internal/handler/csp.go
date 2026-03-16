package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-json"
	goredis "github.com/redis/go-redis/v9"

	"github.com/meysam81/vigil/internal/constants"
	"github.com/meysam81/vigil/internal/httperr"
)

func (h *Handler) HandleReport(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	contentType = strings.ToLower(contentType)

	switch contentType {
	case "application/reports+json", "application/csp-report", "application/json":
		// valid content types
	default:
		h.log.Error().Str("content_type", contentType).Msg("invalid content type rejected")
		httperr.BadRequest(w, httperr.MsgBadContentType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.config.Server.MaxBodySize)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error().Err(err).Msg("failed reading request body")
		httperr.RequestTooLarge(w)
		return
	}

	if !json.Valid(body) {
		h.log.Error().Msg("invalid JSON in request body")
		httperr.BadRequest(w, httperr.MsgInvalidBody)
		return
	}

	now := time.Now()
	reports, err := splitReports(body)
	if err != nil {
		h.log.Error().Err(err).Msg("failed splitting batched reports")
		httperr.BadRequest(w, httperr.MsgInvalidBody)
		return
	}

	pipe := h.redis.Pipeline()
	for _, single := range reports {
		key, err := reportKey(now)
		if err != nil {
			h.log.Error().Err(err).Msg("failed generating report key")
			httperr.Internal(w)
			return
		}
		pipe.Set(r.Context(), key, []byte(single), h.config.Redis.KeyTTL)
		pipe.ZAdd(r.Context(), constants.TimelineKey, goredis.Z{Score: float64(now.UnixNano()), Member: key})
	}
	if _, err := pipe.Exec(r.Context()); err != nil {
		h.log.Error().Err(err).Msg("failed saving body to redis")
		httperr.Internal(w)
		return
	}

	h.log.Info().Int("count", len(reports)).Msg("stored csp violation report(s)")
	w.WriteHeader(http.StatusNoContent)
}

// splitReports handles both single-object and batched array payloads.
// The Reporting API (W3C) allows browsers to send an array of reports
// in a single POST. This splits arrays into individual JSON objects.
func splitReports(body []byte) ([]json.RawMessage, error) {
	trimmed := bytes.TrimLeft(body, " \t\r\n")
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty body")
	}
	if trimmed[0] == '[' {
		var items []json.RawMessage
		if err := json.Unmarshal(trimmed, &items); err != nil {
			return nil, fmt.Errorf("unmarshaling report array: %w", err)
		}
		if len(items) == 0 {
			return nil, fmt.Errorf("empty report array")
		}
		return items, nil
	}
	return []json.RawMessage{trimmed}, nil
}

func reportKey(now time.Time) (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return fmt.Sprintf("csp:%d:%s", now.UnixNano(), hex.EncodeToString(b[:])), nil
}
