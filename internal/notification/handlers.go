package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/riazahmedshah/go-booking/internal/config"
	"github.com/riazahmedshah/go-booking/internal/lib/email"
)

func (n *NotificationService) InitHandlers(config *config.Config) {
	n.emailClient = email.NewSMTPClient(config)
}

func (n *NotificationService) handleBookingCompletion(ctx context.Context, t *asynq.Task) error {
	var p BookingCompletionTask

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		slog.Error("Failed to unmarshal payload", "error", err)
		return fmt.Errorf("failed to booking success email payload: %w", err)
	}

	slog.Info("Sending booking completion email", "booking_id", p.BookingID, "user_id", p.UserID)
	// Check UserRepo
	if n.userRepo == nil {
		slog.Error("userRepo is nil! Did you call SetUserRepo()?")
		return fmt.Errorf("userRepo is nil")
	}
	userEmail, err := n.userRepo.GetUserEmail(ctx, p.UserID)
	if err != nil {
		slog.Error("Failed to resolve user email", "user_id", p.UserID, "error", err)
		return fmt.Errorf("failed to resolve user email for user %s: %w", p.UserID, err)
	}

	// 2. Check SMTP Client
	if n.emailClient == nil {
		slog.Error("emailClient is nil! Did you call InitHandlers()?")
		return fmt.Errorf("emailClient is nil")
	}

	if err := n.emailClient.SendConfirmationEmail(
		userEmail,
		p.BookingID,
		// p.PropertyName,
		// p.StartDate,
		// p.EndDate,
		// p.Address,
		// p.TotalMembers,
		p.TotalPrice,
	); err != nil {
		slog.Error("FAILED TO SEND EMAIL VIA SMTP", "error", err)
		return err
		// return fmt.Errorf("failed to send booking completion email: %w", err)
	}

	slog.Info("Booking completion email sent successfully", "booking_id", p.BookingID, "userEmail", userEmail)
	return nil
}

func (n *NotificationService) HandleSendOTP(email string, otp int) error {
	slog.Info("Sending OTP", "email", email)
	return n.emailClient.SendOTP(email, otp)
}
