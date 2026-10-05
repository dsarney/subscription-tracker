-- +goose Up
CREATE TABLE subscriptions (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    frequency   TEXT NOT NULL CHECK (frequency IN ('monthly', 'yearly')),
    status      TEXT NOT NULL CHECK (status IN ('active', 'inactive')),
    auto_renew  BOOLEAN NOT NULL DEFAULT true
);

-- +goose Down
DROP TABLE IF EXISTS subscriptions;