package handler_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/handler"
	"github.com/meysam81/vigil/internal/logger"
)

var (
	testHandler *handler.Handler
	testRouter  *chi.Mux
)

func TestMain(m *testing.M) {
	cfg := &config.Config{
		LogLevel: "error",
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	log := logger.NewLogger("error", true)

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if _, err := redisClient.Ping(context.TODO()).Result(); err != nil {
		// Skip tests when Redis is not available instead of panicking
		os.Exit(0)
	}

	testHandler = handler.New(redisClient, log, cfg)
	testRouter = chi.NewRouter()
	testRouter.Post("/", testHandler.HandleReport)

	os.Exit(m.Run())
}

func TestPostCSPViolationReportURI(t *testing.T) {
	var body bytes.Buffer
	body.WriteString(`{
  "csp-report": {
    "blocked-uri": "http://example.com/css/style.css",
    "disposition": "report",
    "document-uri": "http://example.com/signup.html",
    "effective-directive": "style-src-elem",
    "original-policy": "default-src 'none'; style-src cdn.example.com; report-uri /_/csp-reports",
    "referrer": "",
    "status-code": 200,
    "violated-directive": "style-src-elem"
  }
}`)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/csp-report; charset=utf-8")

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)

	if recorder.Result().StatusCode != http.StatusNoContent {
		r, _ := io.ReadAll(recorder.Body)
		t.Fatalf("expected %d, got %d: %s", http.StatusNoContent, recorder.Result().StatusCode, string(r))
	}
}

func TestPostCSPViolationReportTo(t *testing.T) {
	var body bytes.Buffer
	body.WriteString(`{
  "age": 53531,
  "body": {
    "blockedURL": "inline",
    "columnNumber": 39,
    "disposition": "enforce",
    "documentURL": "https://example.com/csp-report",
    "effectiveDirective": "script-src-elem",
    "lineNumber": 121,
    "originalPolicy": "default-src 'self'; report-to csp-endpoint-name",
    "referrer": "https://www.google.com/",
    "sample": "console.log(\"lo\")",
    "sourceFile": "https://example.com/csp-report",
    "statusCode": 200
  },
  "type": "csp-violation",
  "url": "https://example.com/csp-report",
  "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36"
}`)

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", "application/reports+json; charset=utf-8")

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)

	if recorder.Result().StatusCode != http.StatusNoContent {
		r, _ := io.ReadAll(recorder.Body)
		t.Fatalf("expected %d, got %d: %s", http.StatusNoContent, recorder.Result().StatusCode, string(r))
	}
}

func TestInvalidContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "text/plain")

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)

	if recorder.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, recorder.Result().StatusCode)
	}
}

func TestMalformedJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{not json`))
	req.Header.Set("Content-Type", "application/csp-report")

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)

	if recorder.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, recorder.Result().StatusCode)
	}
}
