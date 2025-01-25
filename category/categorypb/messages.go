package categorypb

import (
	"github.com/jongyunha/lunchbox/internal/registry"
	"github.com/jongyunha/lunchbox/internal/registry/serdes"
)

const (
	CategoryAggregateChannel = "lunchbox.categories.events.Category"

	CategoryRegisteredEvent = "categoryapi.CategoryRegistered"
)

func Registrations(reg registry.Registry) error {
	serde := serdes.NewProtoSerde(reg)

	// Category events
	if err := serde.Register(&CategoryRegistered{}); err != nil {
		return err
	}

	return nil
}

func (*CategoryRegistered) Key() string { return CategoryRegisteredEvent }
