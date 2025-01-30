-- +goose Up
CREATE OR REPLACE FUNCTION created_at_trigger()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.created_at IS NULL THEN
    NEW.created_at := NOW();
  ELSE
    NEW.created_at := OLD.created_at;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION updated_at_trigger()
RETURNS TRIGGER AS $$
BEGIN
  IF row(NEW.*) IS DISTINCT FROM row(OLD.*) THEN
    NEW.updated_at := NOW();
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;


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
  ON restaurants.restaurants
  FOR EACH ROW EXECUTE PROCEDURE created_at_trigger();

CREATE TRIGGER updated_at_restaurants_trgr
  BEFORE UPDATE
  ON restaurants.restaurants
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
  ON restaurants.snapshots
  FOR EACH ROW EXECUTE PROCEDURE updated_at_trigger();


CREATE TABLE restaurants.inbox (
  id          text        NOT NULL,
  name        text        NOT NULL,
  subject     text        NOT NULL,
  data        bytea       NOT NULL,
  metadata    bytea       NOT NULL,
  sent_at     timestamptz NOT NULL,
  received_at timestamptz NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE restaurants.outbox (
  id           text  NOT NULL,
  name         text  NOT NULL,
  subject      text  NOT NULL,
  data         bytea NOT NULL,
  metadata     bytea       NOT NULL,
  sent_at      timestamptz NOT NULL,
  published_at timestamptz,
  PRIMARY KEY (id)
);

CREATE INDEX restaurants_unpublished_idx ON restaurants.outbox (published_at) WHERE published_at IS NULL;
