package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/riazahmedshah/go-booking/internal/errs"
	"github.com/riazahmedshah/go-booking/internal/lib/utils"
	"github.com/riazahmedshah/go-booking/internal/server"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	server   *server.Server
	userRepo *UserRepository
}

var (
	msgCreateUserFailed     = "failed to create user"
	msgLoginFailed          = "failed to login user"
	msgGetCurrentUserFailed = "failed to get current user"
)

func NewUserService(server *server.Server, ur *UserRepository) *UserService {
	return &UserService{
		server:   server,
		userRepo: ur,
	}
}

func (us *UserService) CreateUser(ctx context.Context, payload *CreateUserPayload) error {
	user, err := us.userRepo.GetUserByEmail(ctx, payload.Email)
	if err == nil && user != nil {
		return errs.ErrDuplicateEmail
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 10)
	if err != nil {
		return errs.New(http.StatusInternalServerError, msgCreateUserFailed, err)
	}

	payload.Password = string(hash)
	if err := us.userRepo.CreateUser(ctx, payload); err != nil {
		return errs.New(http.StatusInternalServerError, msgCreateUserFailed, err)
	}
	return nil
}

func (us *UserService) Login(ctx context.Context, payload *LoginPayload) (string, error) {
	exixtingUser, err := us.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrUserNotFound
		}

		return "", errs.New(http.StatusInternalServerError, msgLoginFailed, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(exixtingUser.Password), []byte(payload.Password))
	if err != nil {
		return "", errs.ErrInvalidPassword
	}

	token, err := utils.GenerateJWTToken(exixtingUser.ID, exixtingUser.Role, []byte(us.server.Config.JWT.SecretKey))
	if err != nil {

		return "", errs.New(http.StatusInternalServerError, msgLoginFailed, err)
	}

	return token, nil
}

func (us *UserService) GetCurrentUser(ctx context.Context, userID string) (*ResponseUserDTO, error) {
	user, err := us.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errs.New(http.StatusInternalServerError, msgGetCurrentUserFailed, err)
	}

	return user, nil
}
