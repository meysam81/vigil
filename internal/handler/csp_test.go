package handler_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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
		Server: config.ServerConfig{
			MaxBodySize: 65536, // 64KB
		},
		Redis: config.RedisConfig{
			Host:   "localhost",
			Port:   6379,
			KeyTTL: 720 * time.Hour,
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

func TestBatchedReports(t *testing.T) {
	payload := `[
  {
    "age": 10,
    "body": {
      "blockedURL": "https://evil.com/tracker.js",
      "disposition": "enforce",
      "documentURL": "https://example.com/page1",
      "effectiveDirective": "script-src-elem"
    },
    "type": "csp-violation",
    "url": "https://example.com/page1"
  },
  {
    "age": 20,
    "body": {
      "blockedURL": "https://evil.com/ads.js",
      "disposition": "report",
      "documentURL": "https://example.com/page2",
      "effectiveDirective": "script-src-elem"
    },
    "type": "csp-violation",
    "url": "https://example.com/page2"
  }
]`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/reports+json")

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)

	if recorder.Result().StatusCode != http.StatusNoContent {
		r, _ := io.ReadAll(recorder.Body)
		t.Fatalf("expected %d, got %d: %s", http.StatusNoContent, recorder.Result().StatusCode, string(r))
	}
}

func TestCORSPreflight(t *testing.T) {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         300,
	}))
	r.Post("/", testHandler.HandleReport)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	if recorder.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected %d for CORS preflight, got %d", http.StatusOK, recorder.Result().StatusCode)
	}
	if acao := recorder.Header().Get("Access-Control-Allow-Origin"); acao != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin=*, got %q", acao)
	}
	if acam := recorder.Header().Get("Access-Control-Allow-Methods"); acam == "" {
		t.Fatal("expected Access-Control-Allow-Methods header")
	}
}

func TestOversizedBody(t *testing.T) {
	smallCfg := &config.Config{
		LogLevel: "error",
		Server: config.ServerConfig{
			MaxBodySize: 64, // 64 bytes
		},
		Redis: config.RedisConfig{
			Host:   "localhost",
			Port:   6379,
			KeyTTL: 720 * time.Hour,
		},
	}
	log := logger.NewLogger("error", true)
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	h := handler.New(redisClient, log, smallCfg)
	r := chi.NewRouter()
	r.Post("/", h.HandleReport)

	// Body larger than 64 bytes.
	largeBody := strings.Repeat(`{"csp-report":{"blocked-uri":"x"}}`, 10)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/csp-report")

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	if recorder.Result().StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected %d, got %d", http.StatusRequestEntityTooLarge, recorder.Result().StatusCode)
	}
}
