-- name: SaveRestaurant :exec
INSERT INTO restaurants.restaurants (id, name) VALUES ($1, $2);