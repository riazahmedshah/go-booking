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
	"github.com/riazahmedshah/go-booking/internal/errs"
	"github.com/riazahmedshah/go-booking/internal/notification"
	"github.com/riazahmedshah/go-booking/internal/server"
	"google.golang.org/api/idtoken"
)

type UserService struct {
	server       *server.Server
	userRepo     *UserRepository
	notification *notification.NotificationService
}

var (
	msgCreateUserFailed = "failed to create user"
	// msgLoginFailed          = "failed to login user"
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

func (us *UserService) VerifyOTP(ctx context.Context, email string, otp int64) (*VerifyOTPResult, error) {
	cmd := us.server.RedisClient.B().Getdel().Key(email).Build()
	otpStr, err := us.server.RedisClient.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, errs.New(http.StatusBadRequest, "otp expired or invalid", err)
		}
		return nil, errs.New(http.StatusInternalServerError, "server error", err)
	}

	otpStored, err := strconv.ParseInt(otpStr, 10, 64)
	if err != nil {
		return nil, errs.New(http.StatusBadRequest, "invalid otp format stored", err)
	}

	if otpStored != otp {
		return nil, errs.New(http.StatusBadRequest, "invalid otp", nil)
	}

	cmdSet := us.server.RedisClient.B().Set().Key("is_verified:" + email).Value("1").Ex(5 * time.Minute).Build()
	if err := us.server.RedisClient.Do(ctx, cmdSet).Error(); err != nil {
		return nil, errs.New(http.StatusInternalServerError, msgVerifyOTPFailed, err)
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
		return nil, errs.New(http.StatusInternalServerError, msgVerifyOTPFailed, err)
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
			return "", errs.New(http.StatusBadRequest, "please veryfy email first", err)
		}
		return "", errs.New(http.StatusInternalServerError, "server error register", err)
	}
	exixtingUser, err := us.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
		return "", errs.New(http.StatusInternalServerError, "server error register get user", err)
	}

	if exixtingUser != nil {
		sessionId, err := CreateSession(ctx, us.server.RedisClient, exixtingUser.ID, exixtingUser.Role)
		if err != nil {
			return "", errs.New(http.StatusInternalServerError, "server error session register", err)
		}
		return sessionId, nil
	}

	user, err := us.userRepo.CreateUser(ctx, payload)
	if err != nil {
		return "", errs.New(http.StatusInternalServerError, msgCreateUserFailed, err)
	}

	sessionId, err := CreateSession(ctx, us.server.RedisClient, user.ID, user.Role)
	if err != nil {
		return "", errs.New(http.StatusInternalServerError, "server error session register", err)
	}

	return sessionId, nil
}

func (us *UserService) Login(ctx context.Context, payload *LoginPayload) (string, error) {

	cmd := us.server.RedisClient.B().Get().Key("is_verified:" + payload.Email).Build()
	err := us.server.RedisClient.Do(ctx, cmd).Error()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return "", errs.New(http.StatusBadRequest, "please veryfy email first", err)
		}
		return "", errs.New(http.StatusInternalServerError, "server error login", err)
	}

	// just key exists

	exixtingUser, err := us.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrUserNotFound
		}

		return "", errs.New(http.StatusInternalServerError, "kya error yaha se aa raha hai??", err)
	}

	sessionId, err := CreateSession(ctx, us.server.RedisClient, exixtingUser.ID, exixtingUser.Role)
	if err != nil {
		return "", errs.New(http.StatusInternalServerError, "server session login error", err)
	}

	return sessionId, nil
}

func (us *UserService) GetCurrentUser(ctx context.Context, userID string) (*ResponseUserDTO, error) {
	user, err := us.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errs.New(http.StatusInternalServerError, msgGetCurrentUserFailed, err)
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
		return "", errs.New(http.StatusInternalServerError, "login with google post err", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)

	var tokenResp GoogleTokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return "", errs.New(http.StatusInternalServerError, "failed to parse token response", err)
	}

	payload, err := idtoken.Validate(ctx, tokenResp.IDToken, us.server.Config.OAuth.GoogleClientID)
	if err != nil {
		return "", errs.New(http.StatusInternalServerError, "failed to validate google id token", err)
	}

	firstName := payload.Claims["given_name"].(string)
	lastName := payload.Claims["family_name"].(string)
	email := payload.Claims["email"].(string)
	isVerified := payload.Claims["email_verified"].(bool)

	user, err := us.userRepo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
		return "", errs.New(http.StatusInternalServerError, "failed to get user by email", err)
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
			return "", errs.New(http.StatusInternalServerError, "failed to create user from google login", err)
		}
	}
	sessionId, err := CreateSession(ctx, us.server.RedisClient, user.ID, user.Role)
	if err != nil {
		return "", errs.New(http.StatusInternalServerError, "failed to create session for existing user", err)
	}
	return sessionId, nil
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
