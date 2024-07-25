package auth

import (
	"context"
	"errors"
	"grpc_service/internal/db"
	"grpc_service/internal/services/auth"
	ssov1 "grpc_service/protos/gen/go/sso"
	"net/mail"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	internalError = status.Error(codes.Internal, "internal error")
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID uint32,
	) (token string, err error)

	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID uint32, err error)

	IsAdmin(ctx context.Context, userID uint32) (bool, error)
}

type serverAPI struct {
	// UnimplementedAuthServer устанавливает заглушки на методы, чтобы при разработке
	// отладка была возможна в случае отсутствия реализации какого-то из методов
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

// Handlers ————————————————————————————————————————
func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	if err := validateLoginReq(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		// TODO: обработка в зависимости от ошибки
		// Неправильный пароль
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, internalError
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if err := validateRegisterReq(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		// TODO: обработка в зависимости от ошибки (то же самое, что и на с.57)
		// Попытался два раза зарегистрироваться
		if errors.Is(err, db.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, internalError
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {

	if err := validateIsAdminReq(req); err != nil {
		return nil, internalError
	}
	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			return nil, status.Error(codes.AlreadyExists, "user not found")
		}
		return nil, internalError
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

//——————————————————————————————————————————————————

// validateHandlers ————————————————————————————————
// TODO: Улучшить валидацию запросов
func validateLoginReq(req *ssov1.LoginRequest) error {
	if err := validateEmail(req.GetEmail()); err != nil {
		return err
	}
	if err := validatePassword(req.GetPassword()); err != nil {
		return err
	}
	return nil
}

func validateRegisterReq(req *ssov1.RegisterRequest) error {
	if err := validateEmail(req.GetEmail()); err != nil {
		return err
	}
	if err := validatePassword(req.GetPassword()); err != nil {
		return err
	}
	return nil
}

func validateIsAdminReq(req *ssov1.IsAdminRequest) error {
	if err := validateUserID(req.GetUserId()); err != nil {
		return err
	}
	return nil
}

//——————————————————————————————————————————————————

// validateAttributes ——————————————————————————————
func validateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return status.Error(codes.InvalidArgument, "email is invalid")
	}
	return nil
}

func validatePassword(password string) error {
	if err := ParsePassword(password); err != nil {
		return err
	}
	return nil
}

func validateUserID(userID uint32) error {
	if err := ParseUserID(userID); err != nil {
		return err
	}
	return nil
}

//———————————————————————————————————————————————————

// Parse ————————————————————————————————————————————
func ParsePassword(password string) error {
	if password == "" {
		return status.Error(codes.InvalidArgument, "password is invalid")
	}
	return nil
}

func ParseUserID(userID uint32) error {
	if userID == 0 {
		return status.Error(codes.InvalidArgument, "userID is invalid")
	}
	return nil
}

//———————————————————————————————————————————————————
