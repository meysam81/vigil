package reporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/logger"
)

func testWebhookSender() *WebhookSender {
	log := logger.NewLogger("error", true)
	cfg := &config.ReporterConfig{
		MaxRetries:    4,
		RetryMinDelay: time.Millisecond,
		RetryMaxDelay: 5 * time.Millisecond,
	}
	return NewWebhookSender(log, cfg)
}

func TestWebhookSender_Success(t *testing.T) {
	var called atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Store(true)
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type: want application/json, got %s", ct)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ws := testWebhookSender()
	err := ws.Send(context.Background(), srv.URL, []byte(`{"test":true}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called.Load() {
		t.Fatal("handler was not called")
	}
}

func TestWebhookSender_Retry(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ws := testWebhookSender()
	err := ws.Send(context.Background(), srv.URL, []byte(`{"test":true}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := attempts.Load()
	if got != 3 {
		t.Fatalf("attempts: want 3, got %d", got)
	}
}

func TestWebhookSender_AllRetriesExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	log := logger.NewLogger("error", true)
	cfg := &config.ReporterConfig{
		MaxRetries:    2,
		RetryMinDelay: time.Millisecond,
		RetryMaxDelay: 2 * time.Millisecond,
	}
	ws := NewWebhookSender(log, cfg)

	err := ws.Send(context.Background(), srv.URL, []byte(`{"test":true}`))
	if err == nil {
		t.Fatal("expected error after all retries exhausted")
	}
}

func TestWebhookSender_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	log := logger.NewLogger("error", true)
	cfg := &config.ReporterConfig{
		MaxRetries:    10,
		RetryMinDelay: time.Second,
		RetryMaxDelay: 2 * time.Second,
	}
	ws := NewWebhookSender(log, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ws.Send(ctx, srv.URL, []byte(`{"test":true}`))
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestJitterDelay(t *testing.T) {
	min := 100 * time.Millisecond
	max := 500 * time.Millisecond

	for range 100 {
		d := jitterDelay(min, max)
		if d < min || d > max {
			t.Fatalf("jitterDelay(%v, %v) = %v, out of range", min, max, d)
		}
	}
}

func TestJitterDelay_EqualBounds(t *testing.T) {
	d := jitterDelay(100*time.Millisecond, 100*time.Millisecond)
	if d != 100*time.Millisecond {
		t.Fatalf("jitterDelay with equal bounds: want 100ms, got %v", d)
	}
}

func TestJitterDelay_MinGreaterThanMax(t *testing.T) {
	d := jitterDelay(500*time.Millisecond, 100*time.Millisecond)
	if d != 500*time.Millisecond {
		t.Fatalf("jitterDelay with min>max: want min (500ms), got %v", d)
	}
}
