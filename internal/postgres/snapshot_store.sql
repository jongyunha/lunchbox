-- name: LoadSnapshot :one
SELECT stream_version, snapshot_name, snapshot_data
FROM restaurants.snapshots
WHERE stream_id = $1 AND stream_name = $2
LIMIT 1;


-- name: SaveSnapshot :exec
INSERT INTO restaurants.snapshots (stream_id, stream_name, stream_version, snapshot_name, snapshot_data)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (stream_id, stream_name) DO UPDATE
SET stream_version = EXCLUDED.stream_version,
    snapshot_name = EXCLUDED.snapshot_name,
    snapshot_data = EXCLUDED.snapshot_data;
