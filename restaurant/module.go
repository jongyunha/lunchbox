package restaurant

import (
	"context"

	"github.com/jongyunha/lunchbox/internal/monolith"
	"github.com/jongyunha/lunchbox/internal/registry"
)

type Module struct{}

func (m *Module) Startup(ctx context.Context, mono monolith.Monolith) (err error) {
	// setup Driven adapters
	reg := registry.New()
	if err = registrations(reg); err != nil {
		return err
	}
	return nil
}

func registrations(_ registry.Registry) error {
	return nil
}
