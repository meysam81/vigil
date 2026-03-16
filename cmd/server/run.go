package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/cors"

	"github.com/meysam81/csp-report-collector/internal/config"
	"github.com/meysam81/csp-report-collector/internal/handler"
	"github.com/meysam81/csp-report-collector/internal/logger"
	"github.com/meysam81/csp-report-collector/internal/middleware"
	iredis "github.com/meysam81/csp-report-collector/internal/redis"
	"github.com/meysam81/x/chimux"
)

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	log = logger.NewLogger(cfg.LogLevel, false)

	if errs := cfg.Validate(); len(errs) > 0 {
		for _, e := range errs {
			log.Error().Msg(e)
		}
		return fmt.Errorf("config validation failed: %s", strings.Join(errs, "; "))
	}

	redisClient, err := iredis.New(ctx, &cfg.Redis)
	if err != nil {
		return fmt.Errorf("connecting to redis: %w", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Error().Err(err).Msg("failed closing redis connection")
		}
	}()

	h := handler.New(redisClient, log, cfg)

	root := chimux.NewChi()
	mw := chimux.NewChi(chimux.WithLoggingMiddleware())
	api := chimux.NewChi(chimux.WithHealthz(), chimux.WithMetrics())

	root.Mount("/", mw)

	if cfg.CORS.Enabled {
		origins := strings.Split(cfg.CORS.AllowedOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		mw.Use(cors.Handler(cors.Options{
			AllowedOrigins: origins,
			AllowedMethods: []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type"},
			MaxAge:         300,
		}))
	}

	mw.Use(middleware.RateLimitMiddleware(redisClient, &cfg.RateLimit, log))
	mw.Mount("/", api)

	api.Post("/", h.ReceiverCSPViolation)
	api.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      root,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info().Msgf("listening on address %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server error: %w", err)
		}
		close(errCh)
	}()

	<-ctx.Done()
	log.Info().Msg("shutdown signal received, draining connections")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	if err := <-errCh; err != nil {
		return err
	}

	log.Info().Msg("shutdown complete")
	return nil
}
