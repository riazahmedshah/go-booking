package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/rueidis"
	"github.com/riazahmedshah/stayz/internal/config"
	"github.com/riazahmedshah/stayz/internal/database"
)

type Server struct {
	Config     *config.Config
	DB         *pgxpool.Pool
	httpServer *http.Server
	// Locker     rueidislock.Locker
	RedisClient rueidis.Client
}

func New(cfg *config.Config) (*Server, error) {
	db, err := database.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	opts, err := rueidis.ParseURL(cfg.Redis.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}
	opts.DisableCache = true
	redisClient, err := rueidis.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Do(ctx, redisClient.B().Ping().Build()).Error(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	server := &Server{
		Config:      cfg,
		DB:          db,
		RedisClient: redisClient,
	}

	return server, nil
}

func (s *Server) SetupHTTPServer(handler http.Handler) {
	s.httpServer = &http.Server{
		Addr:        ":" + s.Config.Server.Port,
		Handler:     handler,
		ReadTimeout: 5 * time.Second,
	}
}

func (s *Server) Start() error {
	if s.httpServer == nil {
		return errors.New("HTTP server not initialized")
	}
	slog.Info("starting server", "port", s.Config.Server.Port)

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	s.DB.Close()
	s.RedisClient.Close()
	return nil
}
