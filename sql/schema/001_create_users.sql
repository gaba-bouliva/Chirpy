-- +goose Up
CREATE TABLE users (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    email   TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL DEFAULT 'unset',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_chirpy_red BOOLEAN NOT NULL DEFAULT FALSE 
);

-- +goose Down
DROP TABLE users;
