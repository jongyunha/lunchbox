package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/jongyunha/lunchbox/restaurants/internal/application"
	"github.com/jongyunha/lunchbox/restaurants/internal/application/commands"
	"github.com/jongyunha/lunchbox/restaurants/restaurantspb"
	"google.golang.org/grpc"
)

type server struct {
	app application.App
	restaurantspb.UnimplementedRestaurantsServiceServer
}

var _ restaurantspb.RestaurantsServiceServer = (*server)(nil)

func RegisterServer(_ context.Context, app application.App, registrar grpc.ServiceRegistrar) error {
	restaurantspb.RegisterRestaurantsServiceServer(registrar, server{app: app})
	return nil
}

func (s server) RegisterRestaurant(ctx context.Context, request *restaurantspb.RegisterRestaurantRequest) (*restaurantspb.RegisterRestaurantResponse, error) {
	restaurantID := uuid.New().String()

	err := s.app.RegisterRestaurant(ctx, commands.RegisterRestaurant{
		ID:   restaurantID,
		Name: request.GetName(),
	})
	if err != nil {
		return nil, err
	}

	return &restaurantspb.RegisterRestaurantResponse{
		Id: restaurantID,
	}, nil
}
