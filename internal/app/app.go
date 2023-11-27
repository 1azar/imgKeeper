package app

import (
	grpcapp "imgKeeper/internal/app/grpc"
	"imgKeeper/internal/service/imgKeeper"
	"imgKeeper/internal/storage/sqlite"
	"log/slog"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	imgKeeperService := imgKeeper.New(log, storage, storage)

	grpcApp := grpcapp.New(log, imgKeeperService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
