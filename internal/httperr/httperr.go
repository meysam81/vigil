package httperr

import (
	"net/http"

	json "github.com/goccy/go-json"
)

const (
	MsgInternal        = "Internal server error"
	MsgBadContentType  = "Bad content-type provided. Only application/reports+json, application/csp-report & application/json is acceptable."
	MsgInvalidBody     = "Invalid request body"
	MsgTooManyRequests = "Too many requests"
	MsgRequestTooLarge = "Request body too large"
)

type errorResponse struct {
	Error string `json:"error"`
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: msg})
}

func Internal(w http.ResponseWriter) {
	WriteError(w, http.StatusInternalServerError, MsgInternal)
}

func BadRequest(w http.ResponseWriter, msg string) {
	WriteError(w, http.StatusBadRequest, msg)
}

func TooManyRequests(w http.ResponseWriter) {
	WriteError(w, http.StatusTooManyRequests, MsgTooManyRequests)
}

func RequestTooLarge(w http.ResponseWriter) {
	WriteError(w, http.StatusRequestEntityTooLarge, MsgRequestTooLarge)
}
