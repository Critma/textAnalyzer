package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Critma/textAnalyzer/analyzer/internal/analyze"
	"github.com/Critma/textAnalyzer/analyzer/internal/config"
	"github.com/Critma/textAnalyzer/analyzer/internal/models"
	"github.com/Critma/textAnalyzer/analyzer/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	} else {
		log.Info().Any("config", cfg).Msg("config loaded")
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     "",
		DB:           0,               // use default DB
		DialTimeout:  5 * time.Second, // connect timeout
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	defer redisClient.Close()

	app := config.Application{
		Config:     cfg,
		Redis:      redisClient,
		HttpClient: &http.Client{Timeout: 5 * time.Second},
	}

	r := gin.Default()
	jobs := make(chan *models.JsonInput, 100)
	routes := routes.New(app, jobs)
	routes.Mount(r.Group("/api"))

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

	go func() {
		// run server in goroutine
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	workersNum := 5
	for i := 0; i < workersNum; i++ {
		// start analyze workers
		go analyze.Worker(jobs, app)
	}

	// Wait for interrupt signal
	<-ctx.Done()
	log.Info().Msg("Shutdown Server...")

	// Shutdown server gracefully
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctxTimeout); err != nil {
		panic(err)
	}

	log.Info().Msg("Server stopped...")
}
