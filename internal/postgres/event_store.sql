-- name: SaveEvent :exec
INSERT INTO events (stream_id, stream_name, stream_version, event_id, event_name, event_data, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: LoadEvents :many
SELECT stream_version, event_id, event_name, event_data, occurred_at
FROM events
WHERE stream_id = $1 AND stream_name = $2 AND stream_version > $3
ORDER BY stream_version ASC;
