package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/cors"
	"golang.org/x/sync/errgroup"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/handler"
	"github.com/meysam81/vigil/internal/logger"
	"github.com/meysam81/vigil/internal/middleware"
	"github.com/meysam81/vigil/internal/migration"
	iredis "github.com/meysam81/vigil/internal/redis"
	"github.com/meysam81/vigil/internal/reporter"
	"github.com/meysam81/x/chimux"
)

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	log := logger.NewLogger(cfg.LogLevel, false)

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	redisClient, err := iredis.New(ctx, &cfg.Redis)
	if err != nil {
		return fmt.Errorf("connecting to redis: %w", err)
	}

	if err := migration.Run(ctx, redisClient, log); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

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

	api.Post("/", h.HandleReport)
	api.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MiB
		Handler:           root,
	}

	g, gctx := errgroup.WithContext(ctx)

	// HTTP server
	g.Go(func() error {
		log.Info().Msgf("listening on address %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	// Reporter (if Slack is configured)
	if cfg.Slack.WebhookURL != "" {
		slackNotifier := reporter.NewSlackNotifier(&cfg.Slack, log)
		rpt := reporter.New(redisClient, log, &cfg.Slack, cfg.Redis.KeyTTL, slackNotifier)
		g.Go(func() error {
			return rpt.Start(gctx)
		})
	}

	// Shutdown watcher
	g.Go(func() error {
		<-gctx.Done()
		log.Info().Msg("shutdown signal received, draining connections")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
		return nil
	})

	err = g.Wait()

	if closeErr := redisClient.Close(); closeErr != nil {
		log.Error().Err(closeErr).Msg("failed closing redis connection")
	}

	log.Info().Msg("shutdown complete")
	return err
}
