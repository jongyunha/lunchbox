package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jongyunha/lunchbox/internal/es"
	"github.com/jongyunha/lunchbox/internal/registry"
)

type SnapshotStore struct {
	es.AggregateStore
	tableName string
	db        *pgxpool.Pool
	registry  registry.Registry
}

var _ es.AggregateStore = (*SnapshotStore)(nil)

func NewSnapshotStore(tableName string, db *pgxpool.Pool, registry registry.Registry) es.AggregateStoreMiddleware {
	snapshots := SnapshotStore{
		tableName: tableName,
		db:        db,
		registry:  registry,
	}

	return func(store es.AggregateStore) es.AggregateStore {
		snapshots.AggregateStore = store
		return snapshots
	}
}

func (s SnapshotStore) Load(ctx context.Context, aggregate es.EventSourcedAggregate) error {
	panic("")
}

func (s SnapshotStore) Save(ctx context.Context, aggregate es.EventSourcedAggregate) error {
	panic("")
}

// TODO use injected & configurable strategies
func (SnapshotStore) shouldSnapshot(aggregate es.EventSourcedAggregate) bool {
	var maxChanges = 3 // low for demonstration; production envs should use higher values 50, 75, 100...
	var pendingVersion = aggregate.PendingVersion()
	var pendingChanges = len(aggregate.Events())

	return pendingVersion >= maxChanges && ((pendingChanges >= maxChanges) ||
		(pendingVersion%maxChanges < pendingChanges) ||
		(pendingVersion%maxChanges == 0))
}

func (s SnapshotStore) table(query string) string {
	return fmt.Sprintf(query, s.tableName)
}
