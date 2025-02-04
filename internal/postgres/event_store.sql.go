// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: event_store.sql

package postgres

import (
	"context"
	"time"
)

const loadEvents = `-- name: LoadEvents :many
SELECT stream_version, event_id, event_name, event_data, occurred_at
FROM restaurants.events
WHERE stream_id = $1 AND stream_name = $2 AND stream_version > $3
ORDER BY stream_version ASC
`

type LoadEventsParams struct {
	StreamID      string `json:"stream_id"`
	StreamName    string `json:"stream_name"`
	StreamVersion int32  `json:"stream_version"`
}

type LoadEventsRow struct {
	StreamVersion int32     `json:"stream_version"`
	EventID       string    `json:"event_id"`
	EventName     string    `json:"event_name"`
	EventData     []byte    `json:"event_data"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (q *Queries) LoadEvents(ctx context.Context, arg LoadEventsParams) ([]LoadEventsRow, error) {
	rows, err := q.db.Query(ctx, loadEvents, arg.StreamID, arg.StreamName, arg.StreamVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LoadEventsRow
	for rows.Next() {
		var i LoadEventsRow
		if err := rows.Scan(
			&i.StreamVersion,
			&i.EventID,
			&i.EventName,
			&i.EventData,
			&i.OccurredAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const saveEvent = `-- name: SaveEvent :exec
INSERT INTO restaurants.events (stream_id, stream_name, stream_version, event_id, event_name, event_data, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

type SaveEventParams struct {
	StreamID      string    `json:"stream_id"`
	StreamName    string    `json:"stream_name"`
	StreamVersion int32     `json:"stream_version"`
	EventID       string    `json:"event_id"`
	EventName     string    `json:"event_name"`
	EventData     []byte    `json:"event_data"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (q *Queries) SaveEvent(ctx context.Context, arg SaveEventParams) error {
	_, err := q.db.Exec(ctx, saveEvent,
		arg.StreamID,
		arg.StreamName,
		arg.StreamVersion,
		arg.EventID,
		arg.EventName,
		arg.EventData,
		arg.OccurredAt,
	)
	return err
}
