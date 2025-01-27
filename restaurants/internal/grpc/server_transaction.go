package grpc

import (
	"context"
	"database/sql"

	"github.com/jongyunha/lunchbox/internal/di"
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
	defer func(tx *sql.Tx) {
		err = s.closeTx(tx, err)
	}(di.Get(ctx, "tx").(*sql.Tx))

	next := server{app: di.Get(ctx, "app").(application.App)}

	return next.RegisterRestaurant(ctx, request)
}

func (s serverTx) closeTx(tx *sql.Tx, err error) error {
	if p := recover(); p != nil {
		_ = tx.Rollback()
		panic(p)
	} else if err != nil {
		_ = tx.Rollback()
		return err
	} else {
		return tx.Commit()
	}
}
