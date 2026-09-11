package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/riazahmedshah/stayz/internal/errs"
	"github.com/riazahmedshah/stayz/internal/server"
)

type UserRepository struct {
	server *server.Server
}

func NewUserRepository(server *server.Server) *UserRepository {
	return &UserRepository{
		server: server,
	}
}

func (ur *UserRepository) CreateUser(ctx context.Context, payload *CreateUserPayload) (*User, error) {
	stmt := `
		INSERT INTO users(
			first_name, last_name, email, is_verified
		) 
		VALUES(
			@first_name, @last_name, @email, @is_verified
		)
		RETURNING *
	`
	rows, err := ur.server.DB.Query(ctx, stmt, pgx.NamedArgs{
		"first_name":  payload.FirstName,
		"last_name":   payload.LastName,
		"email":       payload.Email,
		"is_verified": payload.IsVerified,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute create user query: %w", err)
	}
	defer rows.Close()

	userItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errs.ErrDuplicateEmail
		}
		return nil, fmt.Errorf("failed to execute create user query for email=%s: %w", payload.Email, err)
	}

	return &userItem, nil
}

func (ur *UserRepository) GetUserByID(ctx context.Context, userID string) (*ResponseUserDTO, error) {
	stmt := `
		SELECT
			id, first_name, last_name, email, role, is_verified
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
			id, first_name, last_name, email, role, password, is_verified
		FROM users
		WHERE 
			email=@email
	`
	rows, err := ur.server.DB.Query(ctx, stmt, pgx.NamedArgs{
		"email": email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get user query: %w", err)
	}

	userItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to collect row from table:users for user with email=%s: %w", email, err)
	}
	return &userItem, nil
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

func (ur *UserRepository) UpdateRole(ctx context.Context, userID string) error {
	stmt := `
		UPDATE users
		SET role=@role
		WHERE id=@userId
	`
	rows, err := ur.server.DB.Query(ctx, stmt, pgx.NamedArgs{
		"userId": userID,
		"role":   "host",
	})
	if err != nil {
		return fmt.Errorf("failed to execute update role query: %w", err)
	}
	defer rows.Close()

	return nil
}
