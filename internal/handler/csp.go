package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	json "github.com/goccy/go-json"

	"github.com/meysam81/vigil/internal/httperr"
	"github.com/meysam81/vigil/internal/model"
)

func (h *Handler) ReceiverCSPViolation(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	contentType = strings.ToLower(contentType)

	var cspreport []byte

	switch contentType {
	case "application/reports+json":
		reportTo := &model.ReportTo{}
		if err := json.NewDecoder(r.Body).Decode(reportTo); err != nil {
			h.log.Error().Err(err).Msg("failed decoding request body")
			httperr.BadRequest(w, httperr.MsgInvalidBody)
			return
		}

		var err error
		cspreport, err = json.Marshal(reportTo)
		if err != nil {
			h.log.Error().Err(err).Msg("failed encoding request body for save")
			httperr.Internal(w)
			return
		}

	case "application/csp-report", "application/json":
		reportURI := &model.ReportURI{}
		if err := json.NewDecoder(r.Body).Decode(reportURI); err != nil {
			h.log.Error().Err(err).Msg("failed decoding request body")
			httperr.BadRequest(w, httperr.MsgInvalidBody)
			return
		}

		var err error
		cspreport, err = json.Marshal(reportURI)
		if err != nil {
			h.log.Error().Err(err).Msg("failed encoding request body for save")
			httperr.Internal(w)
			return
		}

	default:
		h.log.Error().Str("content_type", contentType).Msg("invalid content type rejected")
		httperr.BadRequest(w, httperr.MsgBadContentType)
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			h.log.Error().Err(err).Msg("failed closing request body")
		}
	}()

	h.log.Info().Bytes("csp_report", cspreport).Msg("received a csp violation report")

	now := fmt.Sprintf("csp:%d", time.Now().UnixNano())
	if _, err := h.redis.Set(r.Context(), now, cspreport, 0).Result(); err != nil {
		h.log.Error().Err(err).Msg("failed saving body to redis")
	}

	w.WriteHeader(http.StatusNoContent)
}
