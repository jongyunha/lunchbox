-- +goose Up
CREATE SCHEMA restaurants;

CREATE TABLE restaurants.restaurants (
  id            text        NOT NULL,
  name          text        NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT NOW(),
  updated_at    timestamptz NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id)
);

CREATE TRIGGER created_at_restaurants_trgr
  BEFORE UPDATE
  ON restaurants
  FOR EACH ROW EXECUTE PROCEDURE created_at_trigger();

CREATE TRIGGER updated_at_restaurants_trgr
  BEFORE UPDATE
  ON restaurants
  FOR EACH ROW EXECUTE PROCEDURE updated_at_trigger();

CREATE TABLE restaurants.events (
  stream_id      text        NOT NULL,
  stream_name    text        NOT NULL,
  stream_version int         NOT NULL,
  event_id       text        NOT NULL,
  event_name     text        NOT NULL,
  event_data     bytea       NOT NULL,
  occurred_at    timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (stream_id, stream_name, stream_version)
);

CREATE TABLE restaurants.snapshots (
  stream_id      text        NOT NULL,
  stream_name    text        NOT NULL,
  stream_version int         NOT NULL,
  snapshot_name  text        NOT NULL,
  snapshot_data  bytea       NOT NULL,
  updated_at     timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (stream_id, stream_name)
);

CREATE TRIGGER updated_at_snapshots_trgr
  BEFORE UPDATE
  ON snapshots
  FOR EACH ROW EXECUTE PROCEDURE updated_at_trigger();

