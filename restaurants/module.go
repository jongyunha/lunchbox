package restaurants

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jongyunha/lunchbox/internal/am"
	"github.com/jongyunha/lunchbox/internal/amotel"
	"github.com/jongyunha/lunchbox/internal/amprom"
	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/jongyunha/lunchbox/internal/di"
	"github.com/jongyunha/lunchbox/internal/es"
	"github.com/jongyunha/lunchbox/internal/jetstream"
	pg "github.com/jongyunha/lunchbox/internal/postgres"
	"github.com/jongyunha/lunchbox/internal/postgresotel"
	"github.com/jongyunha/lunchbox/internal/registry"
	"github.com/jongyunha/lunchbox/internal/registry/serdes"
	"github.com/jongyunha/lunchbox/internal/system"
	"github.com/jongyunha/lunchbox/internal/tm"
	"github.com/jongyunha/lunchbox/restaurants/internal/application"
	"github.com/jongyunha/lunchbox/restaurants/internal/constants"
	"github.com/jongyunha/lunchbox/restaurants/internal/domain"
	"github.com/jongyunha/lunchbox/restaurants/internal/grpc"
	"github.com/jongyunha/lunchbox/restaurants/internal/handlers"
	"github.com/jongyunha/lunchbox/restaurants/internal/postgres"
	"github.com/jongyunha/lunchbox/restaurants/internal/rest"
	"github.com/jongyunha/lunchbox/restaurants/restaurantspb"
	"github.com/rs/zerolog"
)

type Module struct{}

func (m *Module) Startup(ctx context.Context, svc system.Service) (err error) {
	return Root(ctx, svc)
}

func Root(ctx context.Context, svc system.Service) (err error) {
	container := di.New()

	// setup Driven adapters
	container.AddSingleton(constants.RegistryKey, func(c di.Container) (any, error) {
		reg := registry.New()
		if err = registrations(reg); err != nil {
			return nil, err
		}
		if err = restaurantspb.Registrations(reg); err != nil {
			return nil, err
		}
		return reg, nil
	})

	stream := jetstream.NewStream(svc.Config().Nats.Stream, svc.JS(), svc.Logger())

	container.AddSingleton(constants.DomainDispatcherKey, func(c di.Container) (any, error) {
		return ddd.NewEventDispatcher[ddd.Event](), nil
	})

	container.AddScoped(constants.DatabaseTransactionKey, func(c di.Container) (any, error) {
		return svc.DB().Begin(context.Background())
	})
	sentCounter := amprom.SentMessagesCounter(constants.ServiceName)
	container.AddScoped(constants.MessagePublisherKey, func(c di.Container) (any, error) {
		tx := postgresotel.Trace(c.Get(constants.DatabaseTransactionKey).(pgx.Tx))
		outboxRestaurants := pg.NewOutboxStore(tx)
		return am.NewMessagePublisher(
			stream,
			amotel.OtelMessageContextInjector(),
			sentCounter,
			tm.OutboxPublisher(outboxRestaurants),
		), nil
	})

	container.AddSingleton(constants.MessageSubscriberKey, func(c di.Container) (any, error) {
		return am.NewMessageSubscriber(
			stream,
			amotel.OtelMessageContextExtractor(),
			amprom.ReceivedMessagesCounter(constants.ServiceName),
		), nil
	})

	container.AddScoped(constants.EventPublisherKey, func(c di.Container) (any, error) {
		return am.NewEventPublisher(
			c.Get(constants.RegistryKey).(registry.Registry),
			c.Get(constants.MessagePublisherKey).(am.MessagePublisher),
		), nil
	})

	container.AddScoped(constants.InboxRestaurantKey, func(c di.Container) (any, error) {
		tx := postgresotel.Trace(c.Get(constants.DatabaseTransactionKey).(*pgxpool.Tx))
		return pg.NewInboxStore(tx), nil
	})

	container.AddScoped(constants.AggregateStoreKey, func(c di.Container) (any, error) {
		tx := postgresotel.Trace(c.Get(constants.DatabaseTransactionKey).(*pgxpool.Tx))
		reg := c.Get(constants.RegistryKey).(registry.Registry)
		return es.AggregateStoreWithMiddleware(
			pg.NewEventStore(constants.ServiceName+".events", tx, reg),
			pg.NewSnapshotStore(constants.ServiceName+".snapshots", tx, reg),
		), nil
	})

	container.AddScoped(constants.RestaurantsRepoKey, func(c di.Container) (any, error) {
		return es.NewAggregateRepository[*domain.Restaurant](
			domain.RestaurantAggregate,
			c.Get(constants.RegistryKey).(registry.Registry),
			c.Get(constants.AggregateStoreKey).(es.AggregateStore),
		), nil
	})

	container.AddScoped(constants.MallRepoKey, func(c di.Container) (any, error) {
		return postgres.NewMallRepository(
			postgresotel.Trace(c.Get(constants.DatabaseTransactionKey).(*pgxpool.Tx)),
		), nil
	})

	container.AddScoped(constants.ApplicationKey, func(c di.Container) (any, error) {
		return application.New(
			c.Get(constants.RestaurantsRepoKey).(es.AggregateRepository[*domain.Restaurant]),
			c.Get(constants.DomainDispatcherKey).(ddd.EventPublisher[ddd.Event]),
		), nil
	})

	container.AddScoped(constants.MallHandlersKey, func(c di.Container) (any, error) {
		return handlers.NewMallHandlers(c.Get(constants.MallRepoKey).(domain.MallRepository)), nil
	})
	container.AddScoped(constants.DomainEventHandlersKey, func(c di.Container) (any, error) {
		return handlers.NewDomainEventHandlers(c.Get(constants.EventPublisherKey).(am.EventPublisher)), nil
	})

	outboxProcessor := tm.NewOutboxProcessor(
		stream,
		pg.NewOutboxStore(svc.DB()),
	)

	// setup Driver adapters
	if err = grpc.RegisterServerTx(container, svc.RPC(), svc.Logger()); err != nil {
		return err
	}
	if err = rest.RegisterGateway(ctx, svc.Mux(), svc.Config().Rpc.Address()); err != nil {
		return err
	}
	if err = rest.RegisterSwagger(svc.Mux()); err != nil {
		return err
	}
	handlers.RegisterMallHandlersTx(container)
	handlers.RegisterDomainEventHandlersTx(container)
	startOutboxProcessor(ctx, outboxProcessor, svc.Logger())
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

func startOutboxProcessor(ctx context.Context, outboxProcessor tm.OutboxProcessor, logger zerolog.Logger) {
	go func() {
		err := outboxProcessor.Start(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("stores outbox processor encountered an error")
		}
	}()
}
