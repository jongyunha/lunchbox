package domain

import (
	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/jongyunha/lunchbox/internal/es"
	"github.com/stackus/errors"
)

const (
	RestaurantAggregate = "restaurants.Restaurant"
)

var (
	ErrRestaurantNameIsBlank = errors.Wrap(errors.ErrBadRequest, "the restaurant name cannot be blank")
)

type Restaurant struct {
	es.Aggregate
	Name string
}

func (r *Restaurant) ApplyEvent(event ddd.Event) error {
	switch payload := event.Payload().(type) {
	case *RestaurantRegistered:
		r.Name = payload.Name
	default:
		return errors.ErrInternal.Msgf("%T received the event %s with unexpected payload %T", r, event.EventName(), payload)
	}

	return nil
}

func (r *Restaurant) InitRestaurant(id, name string) (ddd.Event, error) {
	if name == "" {
		return nil, ErrRestaurantNameIsBlank
	}
	//restaurant := NewRestaurant(id)

	r.AddEvent(RestaurantRegisteredEvent, &RestaurantRegistered{
		Name: name,
	})

	return ddd.NewEvent(RestaurantRegisteredEvent, r), nil
}

func (Restaurant) Key() string {
	return RestaurantAggregate
}
