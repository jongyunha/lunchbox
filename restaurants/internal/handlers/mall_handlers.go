package handlers

import (
	"context"
	"time"

	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/jongyunha/lunchbox/internal/di"
	"github.com/jongyunha/lunchbox/internal/errorsotel"
	"github.com/jongyunha/lunchbox/restaurants/internal/constants"
	"github.com/jongyunha/lunchbox/restaurants/internal/domain"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type MallHandlers[T ddd.Event] struct {
	mall domain.MallRepository
}

var _ ddd.EventHandler[ddd.Event] = (*MallHandlers[ddd.Event])(nil)

func NewMallHandlers(mall domain.MallRepository) *MallHandlers[ddd.Event] {
	return &MallHandlers[ddd.Event]{
		mall: mall,
	}
}

func (h MallHandlers[T]) HandleEvent(ctx context.Context, event T) (err error) {
	span := trace.SpanFromContext(ctx)
	defer func(started time.Time) {
		if err != nil {
			span.AddEvent(
				"Encountered an error handling mall event",
				trace.WithAttributes(errorsotel.ErrAttrs(err)...),
			)
		}
		span.AddEvent("Handled mall event", trace.WithAttributes(
			attribute.Int64("TookMS", time.Since(started).Milliseconds()),
		))
	}(time.Now())

	switch event.EventName() {
	case domain.RestaurantRegisteredEvent:
		return h.onRestaurantRegistered(ctx, event)
	}
	return nil
}

func (h MallHandlers[T]) onRestaurantRegistered(ctx context.Context, event ddd.Event) error {
	payload := event.Payload().(*domain.Restaurant)
	return h.mall.RegisterRestaurant(ctx, event.ID(), payload.Name)
}

func RegisterMallHandlers(mallHandlers ddd.EventHandler[ddd.Event], subscriber ddd.EventSubscriber[ddd.Event]) {
	subscriber.Subscribe(mallHandlers, domain.RestaurantRegisteredEvent)
}

func RegisterMallHandlersTx(container di.Container) {
	handlers := ddd.EventHandlerFunc[ddd.Event](func(ctx context.Context, event ddd.Event) error {
		mallHandlers := di.Get(ctx, constants.MallHandlersKey).(ddd.EventHandler[ddd.Event])

		return mallHandlers.HandleEvent(ctx, event)
	})

	subscriber := container.Get(constants.DomainDispatcherKey).(*ddd.EventDispatcher[ddd.Event])
	RegisterMallHandlers(handlers, subscriber)
}
