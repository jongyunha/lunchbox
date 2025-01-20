package restaurantspb

import (
	"github.com/jongyunha/lunchbox/internal/registry"
	"github.com/jongyunha/lunchbox/internal/registry/serdes"
)

const (
	RestaurantAggregateChannel = "lunchbox.restaurant.events.Restaurant"

	RestaurantRegisteredEvent = "restaurantsapi.RestaurantRegistered"
)

func Registrations(reg registry.Registry) error {
	serde := serdes.NewProtoSerde(reg)

	// Restaurant events
	if err := serde.Register(&RestaurantRegistered{}); err != nil {
		return err
	}

	return nil
}

func (*RestaurantRegistered) Key() string {
	return RestaurantRegisteredEvent
}
