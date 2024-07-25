package auth

import (
	"context"
	"errors"
	"fmt"
	"grpc_service/internal/db"
	"grpc_service/internal/domain/models"
	"grpc_service/internal/jwt"
	"grpc_service/internal/slogger"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	emptyValue = 0
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (userID uint32, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID uint32) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID uint32) (models.App, error)
}

// New returns a new istance of Auth service
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

// Login checks if User with given credentials exists in the system and returns token
// If User exist, but password is incorrect, returns error
// If User doesn`t exist, returns error
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID uint32,
) (token string, err error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to login user")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			log.Warn("user not found", slogger.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get user", slogger.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Info("invalid credentials", slogger.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	// ?-Где хранить приватный ключ приложения
	// БД
	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err = jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", slogger.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	return token, nil
}

// RegisterNewUser registers new user in the system and returns userID
// If User with given email already exists, returns error
func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (userID uint32, err error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", slogger.Err(err))
		return emptyValue, fmt.Errorf("%s: %w", op, err)
	}

	userID, err = a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", slogger.Err(err))
		return emptyValue, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return userID, nil
}

// IsAdmin checks if User is admin
func (a *Auth) IsAdmin(
	ctx context.Context,
	userID uint32,
) (isAdmin bool, err error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Any("userID", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err = a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("IsAdmin", isAdmin))

	return isAdmin, nil
}
