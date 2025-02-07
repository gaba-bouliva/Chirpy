-- +goose Up
CREATE TABLE users (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    email   TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE users;
