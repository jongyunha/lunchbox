package crawling

import (
	"context"

	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/jongyunha/lunchbox/internal/monolith"
)

type Module struct{}

func (m *Module) Startup(ctx context.Context, mono monolith.Monolith) error {
	domainDispatcher := ddd.NewEventDispatcher[ddd.Event]()
	return nil
}
