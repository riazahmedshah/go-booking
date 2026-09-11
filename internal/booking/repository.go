package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riazahmedshah/stayz/internal/errs"
	"github.com/riazahmedshah/stayz/internal/server"
)

type BookingRepository struct {
	server *server.Server
}

func NewBookingRepository(server *server.Server) *BookingRepository {
	return &BookingRepository{
		server: server,
	}
}

func (r *BookingRepository) CreateBooking(ctx context.Context, userID string, payload *CreateBookingPayload) (*Booking, error) {
	stmt := `
		INSERT INTO bookings (
			user_id, 
			property_id, 
			total_price, 
			check_in,
			check_out
		)
		SELECT
			@user_id, 
			@property_id, 
			@total_price, 
			@check_in,
			@check_out
		FROM property_availability pa
		WHERE 
			pa.property_id = @property_id
			AND pa.date BETWEEN @check_in AND @check_out
			AND pa.is_available = true
		GROUP BY pa.property_id
		-- Dynamic check: Available rows ka count exact checkin-checkout days ke equal hona chahiye!
		HAVING COUNT(*) = (@check_out::date - @check_in::date + 1)
		RETURNING *
	`

	rows, err := r.server.DB.Query(ctx, stmt, pgx.NamedArgs{
		"user_id":     userID,
		"property_id": payload.PropertyID,
		"total_price": payload.TotalPrice,
		"check_in":    payload.CheckIn,
		"check_out":   payload.CheckOut,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute create booking query for user_id=%v property_id=%v: %w", userID, *payload.PropertyID, err)
	}
	defer rows.Close()

	bookingItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Booking])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrPropertyUnavailable
		}
		return nil, fmt.Errorf("failed to collect row from table:bookings for user_id=%v property_id=%v: %w", userID, *payload.PropertyID, err)
	}

	return &bookingItem, nil
}

func (r *BookingRepository) CreateIdempotencyKey(ctx context.Context, idemKey string, bookingId string) (*IdempotencyKey, error) {
	stmt := `
		INSERT INTO idempotency_keys (
			key, 
			booking_id
		)
		VALUES (
			@key, 
			@booking_id
		)
		RETURNING *
	`

	rows, err := r.server.DB.Query(ctx, stmt, pgx.NamedArgs{
		"key":        idemKey,
		"booking_id": bookingId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute create idempotency key query for key=%v booking_id=%v: %w", idemKey, bookingId, err)
	}
	defer rows.Close()

	idemKeyItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[IdempotencyKey])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:idempotency_keys for key=%v booking_id=%v: %w", idemKey, bookingId, err)
	}

	return &idemKeyItem, nil
}

func (r *BookingRepository) ConfirmBooking(ctx context.Context, tx pgx.Tx, bookingID string) (*Booking, error) {
	stmt := `
		UPDATE bookings
		SET status = @status
		WHERE id = @id
		RETURNING *
	`
	rows, err := tx.Query(ctx, stmt, pgx.NamedArgs{
		"status": "confirmed",
		"id":     bookingID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute confirm booking query for id=%s: %w", bookingID, err)
	}

	data, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Booking])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:bookings for id=%s: %w", bookingID, err)
	}

	return &data, nil
}

func (r *BookingRepository) FinalizeIdempotencyKey(ctx context.Context, tx pgx.Tx, key string) error {
	stmt := `
		UPDATE idempotency_keys
		SET is_finalized = true
		WHERE key = @key
	`
	_, err := tx.Exec(ctx, stmt, pgx.NamedArgs{
		"key": key,
	})
	if err != nil {
		return fmt.Errorf("failed to execute finalize idempotency key query for key=%v: %w", key, err)
	}
	return nil
}

func (r *BookingRepository) GetIdempotencyKeyWithLock(ctx context.Context, tx pgx.Tx, key string) (*IdempotencyKey, error) {
	stmt := `
		SELECT 
			id, 
			key, 
			booking_id, 
			is_finalized,
			created_at,
			updated_at
		FROM 
			idempotency_keys
		WHERE 
			key = @key
		FOR UPDATE
	`
	rows, err := tx.Query(ctx, stmt, pgx.NamedArgs{
		"key": key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get idempotency key with lock query for key=%s: %w", key, err)
	}
	defer rows.Close()

	idemData, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[IdempotencyKey])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:idempotency_keys for key=%v: %w", key, err)
	}

	return &idemData, nil
}

func (r *BookingRepository) UpdatePropertyAvailability(ctx context.Context, tx pgx.Tx, propertyID string, bookingID string, checkIn, checkOut time.Time) error {
	stmt := `
		UPDATE property_availability
		SET 
			booking_id = @booking_id, 
			is_available = false,
			updated_at = NOW()
		WHERE 
			property_id = @property_id 
			AND date BETWEEN @check_in AND @check_out
			AND is_available = true
	`

	_, err := tx.Exec(ctx, stmt, pgx.NamedArgs{
		"property_id": propertyID,
		"booking_id":  bookingID,
		"check_in":    checkIn,
		"check_out":   checkOut,
	})

	if err != nil {
		return fmt.Errorf("failed to execute update property_availability query for property_id=%s booking_id=%s: %w", propertyID, bookingID, err)
	}
	return nil
}
