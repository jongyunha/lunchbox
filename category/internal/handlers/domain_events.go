package handlers

import (
	"context"

	"github.com/jongyunha/lunchbox/category/categorypb"
	"github.com/jongyunha/lunchbox/category/internal/domain"
	"github.com/jongyunha/lunchbox/internal/am"
	"github.com/jongyunha/lunchbox/internal/ddd"
)

type domainHandlers[T ddd.AggregateEvent] struct {
	publisher am.MessagePublisher[ddd.Event]
}

var _ ddd.EventHandler[ddd.AggregateEvent] = (*domainHandlers[ddd.AggregateEvent])(nil)

func NewDomainEventHandlers(publisher am.MessagePublisher[ddd.Event]) *domainHandlers[ddd.AggregateEvent] {
	return &domainHandlers[ddd.AggregateEvent]{
		publisher: publisher,
	}
}

func RegisterDomainEventHandlers(eventHandlers ddd.EventHandler[ddd.AggregateEvent], domainSubscriber ddd.EventSubscriber[ddd.AggregateEvent]) {
	domainSubscriber.Subscribe(
		eventHandlers,
		domain.CategoryRegisteredEvent,
	)

}

func (d domainHandlers[T]) HandleEvent(ctx context.Context, event T) error {
	switch event.EventName() {
	case domain.CategoryRegisteredEvent:
		return d.onCategoryRegistered(ctx, event)
	}

	return nil
}

func (d domainHandlers[T]) onCategoryRegistered(ctx context.Context, event ddd.AggregateEvent) error {
	payload := event.Payload().(domain.CategoryRegistered)
	return d.publisher.Publish(ctx, categorypb.CategoryAggregateChannel,
		ddd.NewEvent(domain.CategoryRegisteredEvent, &categorypb.CategoryRegistered{
			Id:   payload.Category.ID(),
			Name: payload.Category.Name,
		}),
	)
}
