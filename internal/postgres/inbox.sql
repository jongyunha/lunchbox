-- name: SaveRestaurantInboxMessage :one
INSERT INTO restaurants.inbox (id, name, subject, data, received_at) VALUES ($1, $2, $3, $4, $5) RETURNING id;