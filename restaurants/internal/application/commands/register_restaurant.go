package commands

import (
	"context"

	"github.com/jongyunha/lunchbox/restaurants/internal/domain"
)

type (
	RegisterRestaurant struct {
		ID   string
		Name string
	}

	RegisterRestaurantHandler struct {
		restaurants domain.RestaurantRepository
	}
)

func NewRegisterRestaurantHandler(restaurants domain.RestaurantRepository) RegisterRestaurantHandler {
	return RegisterRestaurantHandler{
		restaurants: restaurants,
	}
}

func (h RegisterRestaurantHandler) RegisterRestaurant(ctx context.Context, cmd RegisterRestaurant) error {
	restaurant, err := domain.RegisterRestaurant(cmd.ID, cmd.Name)
	if err != nil {
		return err
	}

	return h.restaurants.Save(ctx, restaurant)
}
