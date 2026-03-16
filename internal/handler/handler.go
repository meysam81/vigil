package handler

import (
	goredis "github.com/redis/go-redis/v9"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/logger"
)

type Handler struct {
	redis  *goredis.Client
	log    *logger.Logger
	config *config.Config
}

func New(redisClient *goredis.Client, log *logger.Logger, cfg *config.Config) *Handler {
	return &Handler{
		redis:  redisClient,
		log:    log,
		config: cfg,
	}
}
