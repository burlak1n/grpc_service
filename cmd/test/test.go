package main

import (
	"errors"
	"grpc_service/internal/slogger"
	"log/slog"
)

func main() {
	log := slogger.SetupLogger("local")
	log.Info("heh")
	log.Error("error")
	log.Warn("wtf")
	log.Info("internal error", slogger.Err(errors.New("error")))
	log.Info("gRPC server is running", slog.String("addr", "10"), slog.Int("port", 5432))
}