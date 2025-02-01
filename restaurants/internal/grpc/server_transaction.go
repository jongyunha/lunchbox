package grpc

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jongyunha/lunchbox/internal/di"
	"github.com/jongyunha/lunchbox/restaurants/internal/application"
	"github.com/jongyunha/lunchbox/restaurants/internal/constants"
	"github.com/jongyunha/lunchbox/restaurants/restaurantspb"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type serverTx struct {
	c di.Container
	restaurantspb.UnimplementedRestaurantsServiceServer
	logger zerolog.Logger
}

var _ restaurantspb.RestaurantsServiceServer = (*serverTx)(nil)

func RegisterServerTx(c di.Container, registrar grpc.ServiceRegistrar, logger zerolog.Logger) error {
	restaurantspb.RegisterRestaurantsServiceServer(
		registrar,
		&serverTx{c: c, logger: logger},
	)

	return nil
}

func (s *serverTx) RegisterRestaurant(ctx context.Context, request *restaurantspb.RegisterRestaurantRequest) (resp *restaurantspb.RegisterRestaurantResponse, err error) {
	ctx = s.c.Scoped(ctx)
	defer func(tx *pgxpool.Tx) {
		err = s.closeTx(ctx, tx, err)
	}(di.Get(ctx, constants.DatabaseTransactionKey).(*pgxpool.Tx))

	next := server{app: di.Get(ctx, "app").(application.App)}

	resp, err = next.RegisterRestaurant(ctx, request)
	if err != nil {
		err = errors.WithStack(err)
		s.logger.Error().Stack().Err(err).Msg("failed to register restaurant")
		return nil, err
	}

	return resp, nil
}

func (s *serverTx) closeTx(ctx context.Context, tx pgx.Tx, err error) error {
	if p := recover(); p != nil {
		_ = tx.Rollback(ctx)
		panic(p)
	} else if err != nil {
		_ = tx.Rollback(ctx)
		return err
	} else {
		return tx.Commit(ctx)
	}
}
