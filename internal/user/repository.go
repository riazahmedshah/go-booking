package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/riazahmedshah/go-booking/internal/errs"
	"github.com/riazahmedshah/go-booking/internal/server"
)

type UserRepository struct {
	server *server.Server
}

func NewUserRepository(server *server.Server) *UserRepository {
	return &UserRepository{
		server: server,
	}
}

func (ur *UserRepository) getUser(ctx context.Context, query string, args pgx.NamedArgs) (*User, error) {
	rows, err := ur.server.DB.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get user query: %w", err)
	}
	defer rows.Close()

	userItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, err
	}
	return &userItem, nil
}

func (ur *UserRepository) CreateUser(ctx context.Context, payload *CreateUserPayload) error {
	stmt := `
		INSERT INTO users(
			first_name, last_name, email, password
		) 
		VALUES(
			@first_name, @last_name, @email, @password
		)
	`
	_, err := ur.server.DB.Exec(ctx, stmt, pgx.NamedArgs{
		"first_name": payload.FirstName,
		"last_name":  payload.LastName,
		"email":      payload.Email,
		"password":   payload.Password,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errs.ErrDuplicateEmail
		}
		return fmt.Errorf("failed to execute create user query for email=%s: %w", payload.Email, err)
	}

	return nil
}

func (ur *UserRepository) GetUserByID(ctx context.Context, userID string) (*ResponseUserDTO, error) {
	stmt := `
		SELECT
			id, first_name, last_name, email, role
		FROM users
		WHERE
			id=@id
	`
	rows, err := ur.server.DB.Query(ctx, stmt, pgx.NamedArgs{
		"id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get user query: %w", err)
	}

	userItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[ResponseUserDTO])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to collect row from table:users for user with id=%s: %w", userID, err)
	}

	return &userItem, nil
}

func (ur *UserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	stmt := `
		SELECT
			id, first_name, last_name, email, password, role
		FROM users
		WHERE 
			email=@email
	`

	user, err := ur.getUser(ctx, stmt, pgx.NamedArgs{"email": email})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to collect row from table:users for user with email=%s: %w", email, err)
	}

	return user, nil
}

func (ur *UserRepository) GetUserEmail(ctx context.Context, userID string) (string, error) {
	stmt := `
		SELECT
			email
		FROM users
		WHERE
			id=@userId
	`

	rows, err := ur.server.DB.Query(ctx, stmt, pgx.NamedArgs{
		"userId": userID,
	})

	if err != nil {
		return "", fmt.Errorf("failed to execute get user query: %w", err)
	}
	defer rows.Close()

	userEmail, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[UserEmail])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrUserNotFound
		}
		return "nil", fmt.Errorf("failed to collect row from table:users for user with id=%s: %w", userID, err)
	}

	return userEmail.Email, nil

}
