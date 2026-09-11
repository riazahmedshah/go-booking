package user

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/rueidis"
	"github.com/riazahmedshah/stayz/internal/errs"
	"github.com/riazahmedshah/stayz/internal/notification"
	"github.com/riazahmedshah/stayz/internal/server"
	"google.golang.org/api/idtoken"
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

func CreateSession(ctx context.Context, client rueidis.Client, userID, role string) (string, error) {
	sid := strings.ToLower(rand.Text())

	// JSON payload
	payload, err := json.Marshal(SessionData{
		UserID: userID,
		Role:   role,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal session data: %w", err)
	}

	// 3. Store in Redis under "session:<sid>"
	key := "session:" + sid
	ttl := time.Hour
	cmd := client.B().Set().Key(key).Value(string(payload)).Ex(ttl).Build()

	if err := client.Do(ctx, cmd).Error(); err != nil {
		return "", fmt.Errorf("failed to save session to redis: %w", err)
	}

	return sid, nil
}

func (us *UserService) SendOTP(ctx context.Context, email string) error {
	max := big.NewInt(1000000)
	otp, err := rand.Int(rand.Reader, max)
	if err != nil {
		slog.Error("crypto/rand error")
		errs.Internal("internal server error", "sendOTP.generateOTP", err)
	}
	// otp := 123456
	if err := us.notification.HandleSendOTP(email, otp.Int64()); err != nil {
		return errs.Internal(msgSendOTPFailed, "sendOTP.HandleSendOTP", err)
	}

	cmd := us.server.RedisClient.B().Set().Key(email).Value(otp.String()).Ex(5 * time.Minute).Build()
	if err := us.server.RedisClient.Do(ctx, cmd).Error(); err != nil {
		return errs.Internal(msgSendOTPFailed, "sendOTP.RedisSet", err)
	}
	return nil
}

func (us *UserService) VerifyOTP(ctx context.Context, email string, otp int64) (*VerifyOTPResult, error) {
	cmd := us.server.RedisClient.B().Getdel().Key(email).Build()
	otpStr, err := us.server.RedisClient.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, errs.ErrInvalidOTP
		}
		return nil, errs.Internal("server error", "verifyOTP.RedisGet", err)
	}

	otpStored, err := strconv.ParseInt(otpStr, 10, 64)
	if err != nil {
		return nil, errs.ErrInvalidOTPFormat
	}

	if otpStored != otp {
		return nil, errs.ErrInvalidOTP
	}

	cmdSet := us.server.RedisClient.B().Set().Key("is_verified:" + email).Value("1").Ex(5 * time.Minute).Build()
	if err := us.server.RedisClient.Do(ctx, cmdSet).Error(); err != nil {
		return nil, errs.Internal(msgVerifyOTPFailed, "verifyOTP.RedisSet", err)
	}

	_, err = us.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			// user does not exists
			return &VerifyOTPResult{
				UserExists: false,
				Email:      email,
			}, nil
		}
		return nil, errs.Internal(msgVerifyOTPFailed, "verifyOTP.GetUserByEmail", err)
	}
	// user exists
	return &VerifyOTPResult{
		UserExists: true,
		Email:      email,
	}, nil
}

func (us *UserService) Register(ctx context.Context, payload *CreateUserPayload) (string, error) {
	cmd := us.server.RedisClient.B().Get().Key("is_verified:" + payload.Email).Build()
	err := us.server.RedisClient.Do(ctx, cmd).Error()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return "", errs.ErrEmailNotVerified
		}
		return "", errs.Internal("server error register", "register.RedisGet", err)
	}
	payload.IsVerified = new(bool)
	*payload.IsVerified = true
	exixtingUser, err := us.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
		return "", errs.Internal("server error register get user", "register.GetUserByEmail", err)
	}

	if exixtingUser != nil {
		sessionId, err := CreateSession(ctx, us.server.RedisClient, exixtingUser.ID, exixtingUser.Role)
		if err != nil {
			return "", errs.Internal("server error session register", "register.CreateSession", err)
		}
		return sessionId, nil
	}

	user, err := us.userRepo.CreateUser(ctx, payload)
	if err != nil {
		return "", errs.Internal(msgCreateUserFailed, "register.CreateUser", err)
	}

	sessionId, err := CreateSession(ctx, us.server.RedisClient, user.ID, user.Role)
	if err != nil {
		return "", errs.Internal(msgCreateUserFailed, "register.CreateSession", err)
	}

	return sessionId, nil
}

func (us *UserService) Login(ctx context.Context, payload *LoginPayload) (string, error) {

	cmd := us.server.RedisClient.B().Get().Key("is_verified:" + payload.Email).Build()
	err := us.server.RedisClient.Do(ctx, cmd).Error()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return "", errs.ErrEmailNotVerified
		}
		return "", errs.Internal(msgLoginFailed, "login.RedisGet", err)
	}

	// just key exists

	exixtingUser, err := us.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrUserNotFound
		}

		return "", errs.Internal(msgLoginFailed, "login.GetUserByEmail", err)
	}

	sessionId, err := CreateSession(ctx, us.server.RedisClient, exixtingUser.ID, exixtingUser.Role)
	if err != nil {
		return "", errs.Internal(msgLoginFailed, "login.CreateSession", err)
	}

	return sessionId, nil
}

func (us *UserService) GetCurrentUser(ctx context.Context, userID string) (*ResponseUserDTO, error) {
	user, err := us.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errs.Internal(msgGetCurrentUserFailed, "getCurrentUser.GetUserByID", err)
	}

	return user, nil
}

func (us *UserService) LoginWithGoogle(ctx context.Context, code string) (string, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {us.server.Config.OAuth.GoogleClientID},
		"client_secret": {us.server.Config.OAuth.GoogleClientSecret},
		"redirect_uri":  {us.server.Config.OAuth.GoogleRedirectURL},
		"grant_type":    {"authorization_code"},
	}

	res, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return "", errs.Internal("login with google post err", "loginWithGoogle.PostForm", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)

	var tokenResp GoogleTokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return "", errs.Internal("failed to parse token response", "loginWithGoogle.Unmarshal", err)
	}

	payload, err := idtoken.Validate(ctx, tokenResp.IDToken, us.server.Config.OAuth.GoogleClientID)
	if err != nil {
		return "", errs.Internal("failed to validate google id token", "loginWithGoogle.Validate", err)
	}

	firstName := payload.Claims["given_name"].(string)
	lastName := payload.Claims["family_name"].(string)
	email := payload.Claims["email"].(string)
	isVerified := payload.Claims["email_verified"].(bool)

	user, err := us.userRepo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
		return "", errs.Internal("failed to get user by email", "loginWithGoogle.GetUserByEmail", err)
	}

	if errors.Is(err, errs.ErrUserNotFound) {
		userPayload := &CreateUserPayload{
			FirstName:  firstName,
			LastName:   &lastName,
			Email:      email,
			IsVerified: &isVerified,
		}

		user, err = us.userRepo.CreateUser(ctx, userPayload)
		if err != nil {
			return "", errs.Internal("failed to create user from google login", "loginWithGoogle.CreateUser", err)
		}
	}
	sessionId, err := CreateSession(ctx, us.server.RedisClient, user.ID, user.Role)
	if err != nil {
		return "", errs.Internal("failed to create session for existing user", "loginWithGoogle.CreateSession", err)
	}
	return sessionId, nil
}

func (us *UserService) UpdateRole(ctx context.Context, userID string) (string, error) {
	cmd := us.server.RedisClient.B().Get().Key("is_verified:" + userID).Build()
	err := us.server.RedisClient.Do(ctx, cmd).Error()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return "errs.ErrEmailNotVerified", nil
		}
		return "", errs.Internal("server error update role", "updateRole.RedisGet", err)
	}
	if err := us.userRepo.UpdateRole(ctx, userID); err != nil {
		return "", errs.Internal("failed to update user role", "updateRole.UpdateRole", err)
	}

	sessionId, err := CreateSession(ctx, us.server.RedisClient, userID, "host")
	if err != nil {
		return "", errs.Internal("failed to create session for updated user", "updateRole.CreateSession", err)
	}

	return sessionId, nil
}
