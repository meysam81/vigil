package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-json"

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

	h.log.Info().RawJSON("csp_report", body).Msg("received a csp violation report")

	key, err := reportKey()
	if err != nil {
		h.log.Error().Err(err).Msg("failed generating report key")
		httperr.Internal(w)
		return
	}

	if _, err := h.redis.Set(r.Context(), key, body, h.config.Redis.KeyTTL).Result(); err != nil {
		h.log.Error().Err(err).Msg("failed saving body to redis")
		httperr.Internal(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func reportKey() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return fmt.Sprintf("csp:%d:%s", time.Now().UnixNano(), hex.EncodeToString(b[:])), nil
}
