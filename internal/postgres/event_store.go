package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jongyunha/lunchbox/internal/es"
	"github.com/jongyunha/lunchbox/internal/registry"
)

type EventStore struct {
	tableName string
	db        *pgxpool.Pool
	registry  registry.Registry
}

var _ es.AggregateStore = (*EventStore)(nil)

func NewEventStore(tableName string, db *pgxpool.Pool, registry registry.Registry) EventStore {
	return EventStore{
		tableName: tableName,
		db:        db,
		registry:  registry,
	}
}

func (e EventStore) Load(ctx context.Context, aggregate es.EventSourcedAggregate) error {
	//TODO implement me
	panic("implement me")
}

func (e EventStore) Save(ctx context.Context, aggregate es.EventSourcedAggregate) error {
	//TODO implement me
	panic("implement me")
}
