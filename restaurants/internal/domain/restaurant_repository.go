package domain

import "context"

type RestaurantRepository interface {
	Load(ctx context.Context, restaurantID string) (*Restaurant, error)
	Save(ctx context.Context, restaurant *Restaurant) error
}
