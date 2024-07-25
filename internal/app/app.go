package app

import (
	"grpc_service/internal/app/grpcapp"
	"grpc_service/internal/db/pg"
	"grpc_service/internal/services/auth"
	"log/slog"
	"time"
)

type App struct {
	GRPCApp *grpcapp.App
}

func New(
	log *slog.Logger,
	gRPCPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	//init storage
	db, err := pg.New(storagePath)

	if err != nil {
		panic(err)
	}

	authService := auth.New(log, db, db, db, tokenTTL)

	gRPCApp := grpcapp.New(log, authService, gRPCPort)

	//init auth service
	return &App{
		GRPCApp: gRPCApp,
	}
}
