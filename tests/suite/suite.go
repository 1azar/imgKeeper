package suite

import (
	"context"
	imgKeeperv1 "github.com/1azar/imgKeeper-api-contracts/gen/go/imgKeeper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"imgKeeper/internal/config"
	"testing"
)

type Suite struct {
	*testing.T
	Cfg             *config.Config
	ImgKeeperClient imgKeeperv1.ImgKeeperClient
}

const configPath

// New creates new test suite.
func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadPath(configPath())

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	cc, err := grpc.DialContext(context.Background(),
		grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials())) // Используем insecure-коннект для тестов
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(cc),
	}
}
