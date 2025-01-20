package restaurants

import (
	"context"

	"github.com/jongyunha/lunchbox/internal/am"
	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/jongyunha/lunchbox/internal/es"
	"github.com/jongyunha/lunchbox/internal/jetstream"
	"github.com/jongyunha/lunchbox/internal/monolith"
	pg "github.com/jongyunha/lunchbox/internal/postgres"
	"github.com/jongyunha/lunchbox/internal/registry"
	"github.com/jongyunha/lunchbox/restaurants/internal/application"
	"github.com/jongyunha/lunchbox/restaurants/internal/domain"
	"github.com/jongyunha/lunchbox/restaurants/internal/grpc"
	"github.com/jongyunha/lunchbox/restaurants/internal/handlers"
	"github.com/jongyunha/lunchbox/restaurants/internal/logging"
	"github.com/jongyunha/lunchbox/restaurants/internal/rest"
)

type Module struct{}

func (m *Module) Startup(ctx context.Context, mono monolith.Monolith) (err error) {
	// setup Driven adapters
	reg := registry.New()
	if err = registrations(reg); err != nil {
		return err
	}
	eventStream := am.NewEventStream(reg, jetstream.NewStream(mono.Config().Nats.Stream, mono.JS(), mono.Logger()))
	domainDispatcher := ddd.NewEventDispatcher[ddd.AggregateEvent]()
	aggregateStore := es.AggregateStoreWithMiddleware(
		pg.NewEventStore("restaurants.events", mono.DB(), reg),
		es.NewEventPublisher(domainDispatcher),
		pg.NewSnapshotStore("restaurants.snapshots", mono.DB(), reg),
	)

	restaurants := es.NewAggregateRepository[*domain.Restaurant](domain.RestaurantAggregate, reg, aggregateStore)

	// setup application
	app := logging.LogApplicationAccess(
		application.New(restaurants),
		mono.Logger(),
	)
	domainEventHandlers := logging.LogEventHandlerAccess[ddd.AggregateEvent](
		handlers.NewDomainEventHandlers(eventStream),
		"DomainEvents", mono.Logger(),
	)
	// setup Driver adapters
	if err = grpc.RegisterServer(ctx, app, mono.RPC()); err != nil {
		return err
	}
	if err = rest.RegisterGateway(ctx, mono.Mux(), mono.Config().Rpc.Address()); err != nil {
		return err
	}
	handlers.RegisterDomainEventHandlers(domainDispatcher, domainEventHandlers)
	return nil
}

func registrations(_ registry.Registry) error {
	return nil
}
