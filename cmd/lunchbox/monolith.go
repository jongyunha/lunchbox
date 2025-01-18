package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jongyunha/lunchbox/internal/config"
	"github.com/jongyunha/lunchbox/internal/monolith"
	"github.com/jongyunha/lunchbox/internal/waiter"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type app struct {
	cfg     config.AppConfig
	dbPool  *pgxpool.Pool
	logger  zerolog.Logger
	modules []monolith.Module
	rpc     *grpc.Server
	mux     *chi.Mux
	waiter  waiter.Waiter
}

func (a *app) Config() config.AppConfig {
	return a.cfg
}

func (a *app) DBPool() *pgxpool.Pool {
	return a.dbPool
}

func (a *app) Logger() zerolog.Logger {
	return a.logger
}

func (a *app) RPC() *grpc.Server {
	return a.rpc
}

func (a *app) Mux() *chi.Mux {
	return a.mux
}

func (a *app) Waiter() waiter.Waiter {
	return a.waiter
}

func (a *app) waitForWeb(ctx context.Context) error {
	webServer := &http.Server{
		Addr:    a.cfg.Web.Address(),
		Handler: a.mux,
	}

	group, gCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		fmt.Printf("web server started; listening at http://localhost%s\n", a.cfg.Web.Port)
		defer fmt.Println("web server shutdown")
		if err := webServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	group.Go(func() error {
		<-gCtx.Done()
		fmt.Println("web server to be shutdown")
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
		defer cancel()
		if err := webServer.Shutdown(ctxWithTimeout); err != nil {
			return err
		}
		return nil
	})

	return group.Wait()
}

func (a *app) waitForRPC(ctx context.Context) error {
	listener, err := net.Listen("tcp", a.cfg.Rpc.Address())
	if err != nil {
		return err
	}

	group, gCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		fmt.Println("rpc server started")
		defer fmt.Println("rpc server shutdown")
		if err := a.RPC().Serve(listener); err != nil && err != grpc.ErrServerStopped {
			return err
		}
		return nil
	})
	group.Go(func() error {
		<-gCtx.Done()
		fmt.Println("rpc server to be shutdown")
		stopped := make(chan struct{})
		go func() {
			a.RPC().GracefulStop()
			close(stopped)
		}()
		timeout := time.NewTimer(a.cfg.ShutdownTimeout)
		select {
		case <-timeout.C:
			// Force it to stop
			a.RPC().Stop()
			return fmt.Errorf("rpc server failed to stop gracefully")
		case <-stopped:
			return nil
		}
	})

	return group.Wait()
}
