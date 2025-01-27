package handlers

import (
	"context"

	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/jongyunha/lunchbox/internal/di"
	"github.com/jongyunha/lunchbox/restaurants/internal/domain"
)

type MallHandlers[T ddd.AggregateEvent] struct {
	mall domain.MallRepository
}

var _ ddd.EventHandler[ddd.AggregateEvent] = (*MallHandlers[ddd.AggregateEvent])(nil)

func NewMallHandlers(mall domain.MallRepository) *MallHandlers[ddd.AggregateEvent] {
	return &MallHandlers[ddd.AggregateEvent]{
		mall: mall,
	}
}

func (h MallHandlers[T]) HandleEvent(ctx context.Context, event T) error {
	switch event.EventName() {
	case domain.RestaurantRegisteredEvent:
		return h.onRestaurantRegistered(ctx, event)
	}
	return nil
}

func (h MallHandlers[T]) onRestaurantRegistered(ctx context.Context, event ddd.AggregateEvent) error {
	payload := event.Payload().(domain.RestaurantRegistered)
	return h.mall.RegisterRestaurant(ctx, event.AggregateID(), payload.Name)
}

func RegisterMallHandlers(mallHandlers ddd.EventHandler[ddd.AggregateEvent], subscriber ddd.EventSubscriber[ddd.AggregateEvent]) {
	subscriber.Subscribe(mallHandlers, domain.RestaurantRegisteredEvent)
}

func RegisterMallHandlersTx(container di.Container) {
	handlers := ddd.EventHandlerFunc[ddd.AggregateEvent](func(ctx context.Context, event ddd.AggregateEvent) error {
		mallHandlers := di.Get(ctx, "mallHandlers").(ddd.EventHandler[ddd.AggregateEvent])

		return mallHandlers.HandleEvent(ctx, event)
	})

	subscriber := container.Get(ddd.DomainDispatcherContainerKey).(*ddd.EventDispatcher[ddd.AggregateEvent])
	RegisterMallHandlers(handlers, subscriber)
}
