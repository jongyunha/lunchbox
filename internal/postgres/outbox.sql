-- name: SaveRestaurantOutboxMessage :one
INSERT INTO restaurants.outbox (id, name, subject, data) VALUES ($1, $2, $3, $4) RETURNING id;

-- name: FindRestaurantUnpublishedOutboxMessages :many
SELECT id, name, subject, data FROM restaurants.outbox WHERE published_at IS NULL LIMIT $1;

-- name: MarkRestaurantOutboxMessageAsPublishedByIDs :exec
UPDATE restaurants.outbox
SET published_at = CURRENT_TIMESTAMP
WHERE id = ANY($1::string[]);
