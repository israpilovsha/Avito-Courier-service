-- +goose Up

CREATE TABLE couriers (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    phone       TEXT NOT NULL UNIQUE,
    status      TEXT NOT NULL, -- например: 'available', 'busy', 'paused'
    created_at  TIMESTAMP DEFAULT now(),
    updated_at  TIMESTAMP DEFAULT now()
);

-- +goose Down

DROP TABLE IF EXISTS couriers;