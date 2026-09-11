package notification

import (
	"context"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/riazahmedshah/stayz/internal/config"
	"github.com/riazahmedshah/stayz/internal/lib/email"
)

type NotificationService struct {
	client      *asynq.Client
	asynqServer *asynq.Server
	userRepo    UserEmailFetcher
	emailClient *email.SMTPClient
}

type UserEmailFetcher interface {
	GetUserEmail(ctx context.Context, userID string) (string, error)
}

func NewNotificationService(cfg *config.Config) *NotificationService {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
	})

	redisOpt, err := redis.ParseURL(cfg.Redis.RedisURL)
	if err != nil {
		slog.Error("failed to parse redis URL for asynq", "error", err)
		os.Exit(1)
	}

	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:      redisOpt.Addr,
			Password:  redisOpt.Password,
			DB:        redisOpt.DB,
			TLSConfig: redisOpt.TLSConfig,
		},
		asynq.Config{
			Concurrency: 10,
		},
	)
	return &NotificationService{
		client:      client,
		asynqServer: server,
	}
}

func (n *NotificationService) SetUserRepo(ur UserEmailFetcher) {
	n.userRepo = ur
}

func (n *NotificationService) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskBookingCompletion, n.handleBookingCompletion)

	slog.Info("Starting background workers...")
	if err := n.asynqServer.Start(mux); err != nil {
		return err
	}
	return nil
}

func (n *NotificationService) Stop() {
	slog.Info("Shutting down background workers...")
	n.asynqServer.Shutdown()
	if err := n.client.Close(); err != nil {
		slog.Error("failed to close notification client", "err", err)
	}
}
