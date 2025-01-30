-- name: SaveRestaurantOutboxMessage :one
INSERT INTO restaurants.outbox (id, name, subject, data, metadata, sent_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;

-- name: FindRestaurantUnpublishedOutboxMessages :many
SELECT id, name, subject, data, metadata, sent_at FROM restaurants.outbox WHERE published_at IS NULL LIMIT $1;

-- name: MarkRestaurantOutboxMessageAsPublishedByIDs :exec
UPDATE restaurants.outbox
SET published_at = CURRENT_TIMESTAMP
WHERE id = ANY($1::text[]);
