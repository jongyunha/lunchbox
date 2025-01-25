package category

import (
	"context"

	"github.com/jongyunha/lunchbox/category/categorypb"
	"github.com/jongyunha/lunchbox/category/internal/application"
	"github.com/jongyunha/lunchbox/category/internal/grpc"
	"github.com/jongyunha/lunchbox/category/internal/handlers"
	"github.com/jongyunha/lunchbox/category/internal/postgres"
	"github.com/jongyunha/lunchbox/category/internal/rest"
	"github.com/jongyunha/lunchbox/internal/am"
	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/jongyunha/lunchbox/internal/jetstream"
	"github.com/jongyunha/lunchbox/internal/monolith"
	"github.com/jongyunha/lunchbox/internal/registry"
)

type Module struct{}

func (m Module) Startup(ctx context.Context, mono monolith.Monolith) (err error) {
	// setup driven adapters
	reg := registry.New()
	if err = categorypb.Registrations(reg); err != nil {
		return err
	}
	stream := jetstream.NewStream(mono.Config().Nats.Stream, mono.JS(), mono.Logger())
	eventStream := am.NewEventStream(reg, stream)
	domainDispatcher := ddd.NewEventDispatcher[ddd.AggregateEvent]()
	categories := postgres.NewCategoryRepository(mono.DB())

	app := application.New(categories, domainDispatcher)
	domainEventHandlers := handlers.NewDomainEventHandlers(eventStream)

	handlers.RegisterDomainEventHandlers(domainEventHandlers, domainDispatcher)

	// setup driver adapters
	if err = grpc.RegisterServer(app, mono.RPC()); err != nil {
		return err
	}
	if err = rest.RegisterGateway(ctx, mono.Mux(), mono.Config().Rpc.Address()); err != nil {
		return err
	}
	if err = rest.RegisterSwagger(mono.Mux()); err != nil {
		return
	}

	return nil
}
