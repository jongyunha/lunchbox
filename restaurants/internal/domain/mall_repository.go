package domain

import "context"

type MallRestaurant struct {
	ID   string
	Name string
}

type MallRepository interface {
	RegisterRestaurant(ctx context.Context, restaurantID, name string) error
	FindByID(ctx context.Context, restaurantID string) (*MallRestaurant, error)
}
