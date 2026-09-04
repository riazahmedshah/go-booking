package user

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/rueidis"
	"github.com/riazahmedshah/go-booking/internal/errs"
	"github.com/riazahmedshah/go-booking/internal/lib/utils"
	"github.com/riazahmedshah/go-booking/internal/notification"
	"github.com/riazahmedshah/go-booking/internal/server"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	server       *server.Server
	userRepo     *UserRepository
	notification *notification.NotificationService
}

var (
	msgCreateUserFailed     = "failed to create user"
	msgLoginFailed          = "failed to login user"
	msgGetCurrentUserFailed = "failed to get current user"
	msgSendOTPFailed        = "failed to send otp"
	msgVerifyOTPFailed      = "failed to verify"
)

func NewUserService(server *server.Server, ur *UserRepository, n *notification.NotificationService) *UserService {
	return &UserService{
		server:       server,
		userRepo:     ur,
		notification: n,
	}
}

func (us *UserService) SendOTP(ctx context.Context, email string) error {
	max := big.NewInt(1000000)
	otp, err := rand.Int(rand.Reader, max)
	if err != nil {
		slog.Error("crypto/rand error")
		errs.New(http.StatusInternalServerError, "server error", err)
	}
	// otp := 123456
	if err := us.notification.HandleSendOTP(email, otp.Int64()); err != nil {
		return errs.New(http.StatusInternalServerError, msgSendOTPFailed, err)
	}

	cmd := us.server.RedisClient.B().Set().Key(email).Value(otp.String()).Ex(5 * time.Minute).Build()
	if err := us.server.RedisClient.Do(ctx, cmd).Error(); err != nil {
		return errs.New(http.StatusInternalServerError, msgSendOTPFailed, err)
	}
	return nil
}

func (us *UserService) VerifyOTP(ctx context.Context, email string, otp int64) error {
	cmd := us.server.RedisClient.B().Get().Key(email).Build()
	otpStr, err := us.server.RedisClient.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return errs.New(http.StatusBadRequest, "otp expired or invalid", err)
		}
		return errs.New(http.StatusInternalServerError, "server error", err)
	}

	otpStored, err := strconv.ParseInt(otpStr, 10, 64)
	if err != nil {
		return errs.New(http.StatusBadRequest, "invalid otp format stored", err)
	}

	if otpStored != otp {
		return errs.New(http.StatusBadRequest, "invalid otp", nil)
	}

	delCmd := us.server.RedisClient.B().Del().Key(email).Build()
	_ = us.server.RedisClient.Do(ctx, delCmd)

	cmdSet := us.server.RedisClient.B().Set().Key("is_verified").Value(email).Ex(time.Minute).Build()
	if err := us.server.RedisClient.Do(ctx, cmdSet).Error(); err != nil {
		return errs.New(http.StatusInternalServerError, msgVerifyOTPFailed, err)
	}
	return nil
}

func (us *UserService) Register(email, name string) error {
	// Register a user
	// Redis GET verified_email:{email}
	// 		- missing: 403 "please verify email first"
	// 		- present? create user row, Redis DEL verified_email:{email}, create session, 201
	return nil
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
