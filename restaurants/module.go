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
	"github.com/jongyunha/lunchbox/internal/registry/serdes"
	"github.com/jongyunha/lunchbox/restaurants/internal/application"
	"github.com/jongyunha/lunchbox/restaurants/internal/domain"
	"github.com/jongyunha/lunchbox/restaurants/internal/grpc"
	"github.com/jongyunha/lunchbox/restaurants/internal/handlers"
	"github.com/jongyunha/lunchbox/restaurants/internal/logging"
	"github.com/jongyunha/lunchbox/restaurants/internal/rest"
	"github.com/jongyunha/lunchbox/restaurants/restaurantspb"
)

type Module struct{}

func (m *Module) Startup(ctx context.Context, mono monolith.Monolith) (err error) {
	// setup Driven adapters
	reg := registry.New()
	if err = registrations(reg); err != nil {
		return err
	}
	if err = restaurantspb.Registrations(reg); err != nil {
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

func registrations(reg registry.Registry) (err error) {
	serde := serdes.NewJsonSerde(reg)

	// Restaurant
	if err = serde.Register(domain.Restaurant{}, func(v any) error {
		restaurant := v.(*domain.Restaurant)
		restaurant.Aggregate = es.NewAggregate("", domain.RestaurantAggregate)
		return nil
	}); err != nil {
		return
	}

	// Restaurant events
	if err = serde.Register(domain.RestaurantRegistered{}); err != nil {
		return
	}

	// Restaurant snapshot
	if err = serde.RegisterKey(domain.RestaurantV1{}.SnapshotName(), domain.RestaurantV1{}); err != nil {
		return
	}

	return
}
