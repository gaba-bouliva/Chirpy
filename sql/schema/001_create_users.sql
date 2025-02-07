-- +goose Up
CREATE TABLE users (
    id SERIAL NOT NULL PRIMARY KEY,
    email   TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE users;
