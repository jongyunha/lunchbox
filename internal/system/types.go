package system

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jongyunha/lunchbox/internal/config"
	"github.com/jongyunha/lunchbox/internal/waiter"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type Service interface {
	Config() config.AppConfig
	DB() *pgxpool.Pool
	JS() nats.JetStreamContext
	Mux() *chi.Mux
	RPC() *grpc.Server
	Waiter() waiter.Waiter
	Logger() zerolog.Logger
}

type Module interface {
	Startup(context.Context, Service) error
}
