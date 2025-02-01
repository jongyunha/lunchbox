package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jongyunha/lunchbox/internal/es"
	"github.com/jongyunha/lunchbox/internal/registry"
)

type SnapshotStore struct {
	es.AggregateStore
	queries   *Queries
	registry  registry.Registry
	tableName string
}

var _ es.AggregateStore = (*SnapshotStore)(nil)

func NewSnapshotStore(tableName string, db DBTX, registry registry.Registry) es.AggregateStoreMiddleware {
	snapshots := &SnapshotStore{
		queries:   New(db),
		registry:  registry,
		tableName: tableName,
	}

	return func(store es.AggregateStore) es.AggregateStore {
		snapshots.AggregateStore = store
		return snapshots
	}
}

func (s *SnapshotStore) Load(ctx context.Context, aggregate es.EventSourcedAggregate) error {
	params := LoadSnapshotParams{
		StreamID:   aggregate.ID(),
		StreamName: aggregate.AggregateName(),
	}

	// 동적으로 테이블 이름 설정
	query := fmt.Sprintf("SELECT stream_version, snapshot_name, snapshot_data FROM %s WHERE stream_id = $1 AND stream_name = $2 LIMIT 1", s.tableName)
	row := s.queries.db.QueryRow(ctx, query, params.StreamID, params.StreamName)

	var snapshotData []byte
	var snapshotName string
	var streamVersion int

	err := row.Scan(&streamVersion, &snapshotName, &snapshotData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.AggregateStore.Load(ctx, aggregate)
		}
		return err
	}

	v, err := s.registry.Deserialize(snapshotName, snapshotData, registry.ValidateImplements((*es.Snapshot)(nil)))
	if err != nil {
		return err
	}

	if err := es.LoadSnapshot(aggregate, v.(es.Snapshot), streamVersion); err != nil {
		return err
	}

	return s.AggregateStore.Load(ctx, aggregate)
}

func (s *SnapshotStore) Save(ctx context.Context, aggregate es.EventSourcedAggregate) error {
	if err := s.AggregateStore.Save(ctx, aggregate); err != nil {
		return err
	}

	if !s.shouldSnapshot(aggregate) {
		return nil
	}

	sser, ok := aggregate.(es.Snapshotter)
	if !ok {
		return fmt.Errorf("%T does not implement es.Snapshotter", aggregate)
	}

	snapshot := sser.ToSnapshot()

	data, err := s.registry.Serialize(snapshot.SnapshotName(), snapshot)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO %s (stream_id, stream_name, stream_version, snapshot_name, snapshot_data) VALUES ($1, $2, $3, $4, $5)", s.tableName)
	params := SaveSnapshotParams{
		StreamID:      aggregate.ID(),
		StreamName:    aggregate.AggregateName(),
		StreamVersion: int32(aggregate.PendingVersion()),
		SnapshotName:  snapshot.SnapshotName(),
		SnapshotData:  data,
	}

	if _, err = s.queries.db.Exec(ctx, query, params.StreamID, params.StreamName, params.StreamVersion, params.SnapshotName, params.SnapshotData); err != nil {
		return err
	}

	return nil
}

// TODO use injected & configurable strategies
func (*SnapshotStore) shouldSnapshot(aggregate es.EventSourcedAggregate) bool {
	var maxChanges = 3 // low for demonstration; production envs should use higher values 50, 75, 100...
	var pendingVersion = aggregate.PendingVersion()
	var pendingChanges = len(aggregate.Events())

	return pendingVersion >= maxChanges && ((pendingChanges >= maxChanges) ||
		(pendingVersion%maxChanges < pendingChanges) ||
		(pendingVersion%maxChanges == 0))
}
