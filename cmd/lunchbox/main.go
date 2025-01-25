package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jongyunha/lunchbox/category"
	"github.com/jongyunha/lunchbox/internal/config"
	"github.com/jongyunha/lunchbox/internal/logger"
	"github.com/jongyunha/lunchbox/internal/monolith"
	"github.com/jongyunha/lunchbox/internal/postgres"
	"github.com/jongyunha/lunchbox/internal/rpc"
	"github.com/jongyunha/lunchbox/internal/waiter"
	"github.com/jongyunha/lunchbox/internal/web"
	"github.com/jongyunha/lunchbox/restaurants"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run() (err error) {
	var cfg config.AppConfig
	cfg, err = config.InitConfig()
	if err != nil {
		return err
	}

	m := app{cfg: cfg}
	dbPool, err := initDbPool(cfg.PG)
	if err != nil {
		return err
	}
	defer func() {
		dbPool.Close()
	}()

	m.nc, err = nats.Connect(cfg.Nats.URL)
	if err != nil {
		return err
	}
	defer m.nc.Close()

	m.js, err = initJetStream(cfg.Nats, m.nc)
	if err != nil {
		return err
	}
	m.dbPool = dbPool
	m.queries = postgres.New(dbPool)
	m.logger = initLogger(cfg)
	m.rpc = initRpc(cfg.Rpc)
	m.mux = initMux(cfg.Web)
	m.waiter = waiter.New(waiter.CatchSignals())

	m.modules = []monolith.Module{
		&restaurants.Module{},
		&category.Module{},
		//&crawling.Module{},
	}

	if err = m.startupModules(); err != nil {
		return err
	}

	webUi := http.FS(web.WebUI)
	m.mux.Mount("/", http.FileServer(webUi))

	fmt.Println("lunchbox is running")
	defer fmt.Println("lunchbox stopped")

	m.Waiter().Add(
		m.waitForWeb,
		m.waitForRPC,
	)

	return m.waiter.Wait()
}

func initDbPool(cfg config.PGConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = time.Duration(cfg.MaxConnLifetime) * time.Second
	poolConfig.MaxConnIdleTime = time.Duration(cfg.MaxConnIdleTime) * time.Second
	poolConfig.HealthCheckPeriod = time.Duration(cfg.HealthCheckPeriod) * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return pool, nil
}

func initRpc(_ rpc.RpcConfig) *grpc.Server {
	server := grpc.NewServer()
	reflection.Register(server)

	return server
}

func initLogger(cfg config.AppConfig) zerolog.Logger {
	return logger.New(logger.LogConfig{
		Environment: cfg.Environment,
		LogLevel:    logger.Level(cfg.LogLevel),
	})
}

func initMux(_ web.WebConfig) *chi.Mux {
	return chi.NewMux()
}

func initJetStream(cfg config.NatsConfig, nc *nats.Conn) (nats.JetStreamContext, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     cfg.Stream,
		Subjects: []string{fmt.Sprintf("%s.>", cfg.Stream)},
	})

	return js, err
}
