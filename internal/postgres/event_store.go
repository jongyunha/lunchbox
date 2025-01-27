package postgres

import (
	"context"
	"fmt"

	"github.com/jongyunha/lunchbox/internal/es"
	"github.com/jongyunha/lunchbox/internal/registry"
)

type EventStore struct {
	tableName string
	queries   *Queries
	registry  registry.Registry
}

var _ es.AggregateStore = (*EventStore)(nil)

func NewEventStore(tableName string, db DBTX, registry registry.Registry) EventStore {
	return EventStore{
		tableName: tableName,
		queries:   New(db),
		registry:  registry,
	}
}

func (s EventStore) Load(ctx context.Context, aggregate es.EventSourcedAggregate) error {
	aggregateID := aggregate.ID()
	aggregateName := aggregate.AggregateName()
	currentVersion := aggregate.Version()

	rows, err := s.queries.LoadEvents(ctx, LoadEventsParams{
		StreamID:      aggregateID,
		StreamName:    aggregateName,
		StreamVersion: int32(currentVersion),
	})
	if err != nil {
		return err
	}

	for _, row := range rows {
		payload, err := s.registry.Deserialize(row.EventName, row.EventData)
		if err != nil {
			return err
		}

		event := aggregateEvent{
			id:         row.EventID,
			name:       row.EventName,
			payload:    payload,
			aggregate:  aggregate,
			version:    int(row.StreamVersion),
			occurredAt: row.OccurredAt,
		}

		if err := es.LoadEvent(aggregate, event); err != nil {
			return err
		}
	}

	return nil
}

func (s EventStore) Save(ctx context.Context, aggregate es.EventSourcedAggregate) (err error) {
	aggregateID := aggregate.ID()
	aggregateName := aggregate.AggregateName()

	for _, event := range aggregate.Events() {
		payloadData, err := s.registry.Serialize(event.EventName(), event.Payload())
		if err != nil {
			return err
		}

		params := SaveEventParams{
			StreamID:      aggregateID,
			StreamName:    aggregateName,
			StreamVersion: int32(event.AggregateVersion()),
			EventID:       event.ID(),
			EventName:     event.EventName(),
			EventData:     payloadData,
			OccurredAt:    event.OccurredAt(),
		}
		if err = s.saveEvent(ctx, params); err != nil {
			return err
		}
	}
	return nil
}

func (s EventStore) saveEvent(ctx context.Context, params SaveEventParams) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (stream_id, stream_name, stream_version, event_id, event_name, event_data, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7);`, s.tableName)

	_, err := s.queries.db.Exec(ctx, query,
		params.StreamID, params.StreamName, params.StreamVersion,
		params.EventID, params.EventName, params.EventData, params.OccurredAt)
	return err
}

func (s EventStore) loadEvents(ctx context.Context, streamID string, streamName string, streamVersion int32) ([]LoadEventsRow, error) {
	query := fmt.Sprintf(`
		SELECT stream_version, event_id, event_name, event_data, occurred_at
		FROM %s
		WHERE stream_id = $1 AND stream_name = $2 AND stream_version > $3
		ORDER BY stream_version ASC;`, s.tableName)

	rows, err := s.queries.db.Query(ctx, query, streamID, streamName, streamVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []LoadEventsRow
	for rows.Next() {
		var row LoadEventsRow
		if err := rows.Scan(&row.StreamVersion, &row.EventID, &row.EventName, &row.EventData, &row.OccurredAt); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, nil
}
