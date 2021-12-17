-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE events (
                               id uuid NOT NULL DEFAULT uuid_generate_v4(),
                               title varchar NOT NULL,
                               start_timestamp timestamp(0) NOT NULL,
                               end_timestamp timestamp(0) NULL,
                               description varchar NULL,
                               notify_before int8 NULL,
                               owner_id varchar NOT NULL,
                               CONSTRAINT events_pk PRIMARY KEY (id)
);
-- +goose StatementEnd
CREATE UNIQUE INDEX events_id_idx ON events (id);
CREATE INDEX events_start_timestamp_idx ON events (start_timestamp);

-- +goose Down
DROP INDEX events_id_idx;
DROP INDEX events_start_timestamp_idx;
DROP TABLE events;




