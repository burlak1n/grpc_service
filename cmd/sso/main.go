package main

import (
	"grpc_service/internal/app"
	"grpc_service/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"grpc_service/internal/slogger/prettyhandler"
	// . "grpc/internal/entities"
	// "grpc/protos/gen/go/sso"
)

func main() {
	// init config
	cfg := config.MustLoad()

	// slog - logger
	log := prettyhandler.SetupLogger(cfg.Env)

	log.Info("starting application...",
		slog.Any("cfg", cfg),
	)

	application := app.New(
		log,
		cfg.GRPC.Port,
		cfg.StoragePath,
		cfg.TokenTTL,
	)

	go func() {
		application.GRPCApp.MustRun()
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.GRPCApp.Stop()
	log.Info("Gracefully stopped")
	// app init

	// Запустить grpc сервис
}
