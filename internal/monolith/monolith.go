package monolith

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

type Monolith interface {
	Config() config.AppConfig
	DB() *pgxpool.Pool
	Logger() zerolog.Logger
	JS() nats.JetStreamContext
	Mux() *chi.Mux
	RPC() *grpc.Server
	Waiter() waiter.Waiter
}

type Module interface {
	Startup(context.Context, Monolith) error
}
