package notification

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/riazahmedshah/go-booking/internal/config"
	"github.com/riazahmedshah/go-booking/internal/lib/email"
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

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Address, Password: cfg.Redis.Password, DB: 0},
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
