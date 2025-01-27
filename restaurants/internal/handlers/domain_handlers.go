package handlers

import (
	"context"

	"github.com/jongyunha/lunchbox/internal/am"
	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/jongyunha/lunchbox/internal/di"
	"github.com/jongyunha/lunchbox/restaurants/internal/domain"
	"github.com/jongyunha/lunchbox/restaurants/restaurantspb"
)

type domainHandlers[T ddd.AggregateEvent] struct {
	publisher am.MessagePublisher[ddd.Event]
}

var _ ddd.EventHandler[ddd.AggregateEvent] = (*domainHandlers[ddd.AggregateEvent])(nil)

func NewDomainEventHandlers(publisher am.MessagePublisher[ddd.Event]) ddd.EventHandler[ddd.AggregateEvent] {
	return &domainHandlers[ddd.AggregateEvent]{
		publisher: publisher,
	}
}

func RegisterDomainEventHandlers(subscriber ddd.EventSubscriber[ddd.AggregateEvent], handlers ddd.EventHandler[ddd.AggregateEvent]) {
	subscriber.Subscribe(handlers,
		domain.RestaurantRegisteredEvent,
	)
}

func RegisterDomainEventHandlersTx(container di.Container) {
	handlers := ddd.EventHandlerFunc[ddd.AggregateEvent](func(ctx context.Context, event ddd.AggregateEvent) error {
		domainHandlers := di.Get(ctx, "domainEventHandlers").(ddd.EventHandler[ddd.AggregateEvent])

		return domainHandlers.HandleEvent(ctx, event)
	})

	subscriber := container.Get(ddd.DomainDispatcherContainerKey).(*ddd.EventDispatcher[ddd.AggregateEvent])
	RegisterDomainEventHandlers(subscriber, handlers)
}

func (d domainHandlers[T]) HandleEvent(ctx context.Context, event T) error {
	switch event.EventName() {
	case domain.RestaurantRegisteredEvent:
		return d.onRestaurantRegistered(ctx, event)
	}
	return nil
}

func (d domainHandlers[T]) onRestaurantRegistered(ctx context.Context, event T) error {
	payload := event.Payload().(domain.RestaurantRegistered)
	return d.publisher.Publish(ctx, restaurantspb.RestaurantAggregateChannel, ddd.NewEvent(
		restaurantspb.RestaurantRegisteredEvent,
		&restaurantspb.RestaurantRegistered{
			Id:   event.AggregateID(),
			Name: payload.Name,
		},
	))
}
