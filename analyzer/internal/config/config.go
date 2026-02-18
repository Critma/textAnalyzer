package config

import (
	"errors"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

type Application struct {
	Config     *Config
	Redis      *redis.Client
	HttpClient *http.Client
}

type Config struct {
	Addr         string
	ReceiverAddr string
	RedisAddr    string
}

const (
	RECEIVER_ADDR = "RECEIVER_ADDR"
	ANALYZER_ADDR = "ANALYZER_ADDR"
	REDIS_ADDR    = "REDIS_ADDR"
)

func Load() (*Config, error) {
	cfg := &Config{}

	addr, ok := os.LookupEnv(ANALYZER_ADDR)
	if !ok {
		return nil, errors.New("failed to get " + RECEIVER_ADDR)
	}
	cfg.Addr = addr

	addr, ok = os.LookupEnv(RECEIVER_ADDR)
	if !ok {
		return nil, errors.New("failed to get " + ANALYZER_ADDR)
	}
	cfg.ReceiverAddr = addr

	addr, ok = os.LookupEnv(REDIS_ADDR)
	if !ok {
		return nil, errors.New("failed to get " + REDIS_ADDR)
	}
	cfg.RedisAddr = addr
	return cfg, nil
}
