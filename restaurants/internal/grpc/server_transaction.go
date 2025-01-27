package grpc

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jongyunha/lunchbox/internal/di"
	pg "github.com/jongyunha/lunchbox/internal/postgres"
	"github.com/jongyunha/lunchbox/restaurants/internal/application"
	"github.com/jongyunha/lunchbox/restaurants/restaurantspb"
	"google.golang.org/grpc"
)

type serverTx struct {
	c di.Container
	restaurantspb.UnimplementedRestaurantsServiceServer
}

var _ restaurantspb.RestaurantsServiceServer = (*serverTx)(nil)

func RegisterServerTx(c di.Container, registrar grpc.ServiceRegistrar) error {
	restaurantspb.RegisterRestaurantsServiceServer(
		registrar,
		serverTx{c: c},
	)

	return nil
}

func (s serverTx) RegisterRestaurant(ctx context.Context, request *restaurantspb.RegisterRestaurantRequest) (resp *restaurantspb.RegisterRestaurantResponse, err error) {
	ctx = s.c.Scoped(ctx)
	defer func(tx pgx.Tx) {
		err = s.closeTx(ctx, tx, err)
	}(di.Get(ctx, pg.TxContainerKey).(pgx.Tx))

	next := server{app: di.Get(ctx, "app").(application.App)}

	return next.RegisterRestaurant(ctx, request)
}

func (s serverTx) closeTx(ctx context.Context, tx pgx.Tx, err error) error {
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
