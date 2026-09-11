package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/riazahmedshah/go-booking/internal/config"
	"github.com/riazahmedshah/go-booking/internal/handler"
	"github.com/riazahmedshah/go-booking/internal/notification"
	"github.com/riazahmedshah/go-booking/internal/repository"
	"github.com/riazahmedshah/go-booking/internal/router"
	"github.com/riazahmedshah/go-booking/internal/server"
	"github.com/riazahmedshah/go-booking/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	// slog.Info("config loaded successfully", "REDIS_URL", cfg.Redis.RedisURL)
	srv, err := server.New(cfg)
	if err != nil {
		slog.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	repos := repository.NewRepositories(srv)
	notificationService := notification.NewNotificationService(cfg)
	notificationService.InitHandlers(cfg)
	notificationService.SetUserRepo(repos.UserRepository)

	if err := notificationService.Start(); err != nil {
		slog.Error("failed to start notification worker", "error", err)
		os.Exit(1)
	}

	service, err := service.NewService(srv, repos, notificationService)
	if err != nil {
		slog.Error("failed to create service", "error", err)
		os.Exit(1)
	}

	handler := handler.NewHandler(srv, service)

	router := router.NewRouter(srv, handler)

	srv.SetupHTTPServer(router)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
	}

	notificationService.Stop()
	slog.Info("notification workers stopped successfully")
}
