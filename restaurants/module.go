package restaurants

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jongyunha/lunchbox/internal/am"
	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/jongyunha/lunchbox/internal/di"
	"github.com/jongyunha/lunchbox/internal/es"
	"github.com/jongyunha/lunchbox/internal/jetstream"
	"github.com/jongyunha/lunchbox/internal/logger"
	"github.com/jongyunha/lunchbox/internal/monolith"
	pg "github.com/jongyunha/lunchbox/internal/postgres"
	"github.com/jongyunha/lunchbox/internal/registry"
	"github.com/jongyunha/lunchbox/internal/registry/serdes"
	"github.com/jongyunha/lunchbox/internal/tm"
	"github.com/jongyunha/lunchbox/restaurants/internal/application"
	"github.com/jongyunha/lunchbox/restaurants/internal/domain"
	"github.com/jongyunha/lunchbox/restaurants/internal/grpc"
	"github.com/jongyunha/lunchbox/restaurants/internal/handlers"
	"github.com/jongyunha/lunchbox/restaurants/internal/logging"
	"github.com/jongyunha/lunchbox/restaurants/internal/postgres"
	"github.com/jongyunha/lunchbox/restaurants/internal/rest"
	"github.com/jongyunha/lunchbox/restaurants/restaurantspb"
	"github.com/rs/zerolog"
)

type Module struct{}

func (m *Module) Startup(ctx context.Context, mono monolith.Monolith) (err error) {
	container := di.New()

	// setup Driven adapters
	container.AddSingleton(registry.ContainerKey, func(c di.Container) (any, error) {
		reg := registry.New()
		if err = registrations(reg); err != nil {
			return nil, err
		}
		if err = restaurantspb.Registrations(reg); err != nil {
			return nil, err
		}
		return reg, nil
	})
	container.AddSingleton(logger.ContainerKey, func(c di.Container) (any, error) {
		return mono.Logger(), nil
	})
	container.AddSingleton(jetstream.ContainerKey, func(c di.Container) (any, error) {
		return jetstream.NewStream(mono.Config().Nats.Stream, mono.JS(), c.Get(logger.ContainerKey).(zerolog.Logger)), nil
	})

	container.AddSingleton(ddd.DomainDispatcherContainerKey, func(c di.Container) (any, error) {
		return ddd.NewEventDispatcher[ddd.AggregateEvent](), nil
	})

	container.AddSingleton(pg.DBContainerKey, func(c di.Container) (any, error) {
		return mono.DB(), nil
	})

	container.AddSingleton(tm.OutboxProcessorContainerKey, func(c di.Container) (any, error) {
		return tm.NewOutboxProcessor(
			c.Get(jetstream.ContainerKey).(am.RawMessageStream),
			pg.NewOutboxStore(c.Get(pg.DBContainerKey).(*pgxpool.Pool)),
		), nil
	})

	container.AddScoped(pg.TxContainerKey, func(c di.Container) (any, error) {
		return mono.DB().Begin(context.Background())
	})

	container.AddScoped("txStream", func(c di.Container) (any, error) {
		tx := c.Get(pg.TxContainerKey).(*pgxpool.Tx)
		outboxStore := pg.NewOutboxStore(tx)
		return am.RawMessageStreamWithMiddleware(
			c.Get(jetstream.ContainerKey).(am.RawMessageStream),
			tm.NewOutboxStreamMiddleware(outboxStore),
		), nil
	})

	container.AddScoped("eventStream", func(c di.Container) (any, error) {
		return am.NewEventStream(
			c.Get(registry.ContainerKey).(registry.Registry),
			c.Get("txStream").(am.RawMessageStream),
		), nil
	})

	container.AddScoped("aggregateStore", func(c di.Container) (any, error) {
		tx := c.Get(pg.TxContainerKey).(*pgxpool.Tx)
		reg := c.Get(registry.ContainerKey).(registry.Registry)
		return es.AggregateStoreWithMiddleware(
			pg.NewEventStore("restaurants.events", tx, reg),
			es.NewEventPublisher(c.Get(ddd.DomainDispatcherContainerKey).(*ddd.EventDispatcher[ddd.AggregateEvent])),
			pg.NewSnapshotStore("restaurants.snapshots", tx, reg),
		), nil
	})

	container.AddScoped("restaurants", func(c di.Container) (any, error) {
		return es.NewAggregateRepository[*domain.Restaurant](
			domain.RestaurantAggregate,
			c.Get(registry.ContainerKey).(registry.Registry),
			c.Get("aggregateStore").(es.AggregateStore),
		), nil
	})

	container.AddScoped("mall", func(c di.Container) (any, error) {
		tx := c.Get(pg.TxContainerKey).(*pgxpool.Tx)
		return postgres.NewMallRepository(tx), nil
	})

	// setup application
	container.AddScoped("app", func(c di.Container) (any, error) {
		return logging.LogApplicationAccess(
			application.New(c.Get("restaurants").(es.AggregateRepository[*domain.Restaurant])),
			c.Get(logger.ContainerKey).(zerolog.Logger),
		), nil
	})

	container.AddScoped("domainEventHandlers", func(c di.Container) (any, error) {
		return logging.LogEventHandlerAccess[ddd.AggregateEvent](
			handlers.NewDomainEventHandlers(c.Get("eventStream").(am.EventStream)),
			"DomainEvents", c.Get(logger.ContainerKey).(zerolog.Logger),
		), nil
	})

	container.AddScoped("mallHandlers", func(c di.Container) (any, error) {
		return logging.LogEventHandlerAccess[ddd.AggregateEvent](
			handlers.NewMallHandlers(c.Get("mall").(postgres.MallRepository)),
			"Mall", c.Get(logger.ContainerKey).(zerolog.Logger),
		), nil
	})

	// setup Driver adapters
	if err = grpc.RegisterServerTx(container, mono.RPC()); err != nil {
		return err
	}
	if err = rest.RegisterGateway(ctx, mono.Mux(), mono.Config().Rpc.Address()); err != nil {
		return err
	}
	if err = rest.RegisterSwagger(mono.Mux()); err != nil {
		return err
	}
	handlers.RegisterMallHandlersTx(container)
	handlers.RegisterDomainEventHandlersTx(container)
	startOutboxProcessor(ctx, container)
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
	return nil
}

func startOutboxProcessor(ctx context.Context, container di.Container) {
	processor := container.Get(tm.OutboxProcessorContainerKey).(tm.OutboxProcessor)
	logger := container.Get(logger.ContainerKey).(zerolog.Logger)
	go func() {
		if err := processor.Start(ctx); err != nil {
			logger.Error().Err(err).Msg("OutboxProcessor failed")
		}
	}()
}
