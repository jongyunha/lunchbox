package commands

import (
	"context"

	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/jongyunha/lunchbox/restaurants/internal/domain"
)

type (
	RegisterRestaurant struct {
		ID   string
		Name string
	}

	RegisterRestaurantHandler struct {
		restaurants domain.RestaurantRepository
		publisher   ddd.EventPublisher[ddd.Event]
	}
)

func NewRegisterRestaurantHandler(restaurants domain.RestaurantRepository, publisher ddd.EventPublisher[ddd.Event]) RegisterRestaurantHandler {
	return RegisterRestaurantHandler{
		restaurants: restaurants,
		publisher:   publisher,
	}
}

func (h RegisterRestaurantHandler) RegisterRestaurant(ctx context.Context, cmd RegisterRestaurant) error {
	restaurant, err := h.restaurants.Load(ctx, cmd.ID)
	if err != nil {
		return err
	}

	event, err := restaurant.InitRestaurant(cmd.ID, cmd.Name)
	if err != nil {
		return err
	}

	err = h.restaurants.Save(ctx, restaurant)
	if err != nil {
		return err
	}

	return h.publisher.Publish(ctx, event)
}
