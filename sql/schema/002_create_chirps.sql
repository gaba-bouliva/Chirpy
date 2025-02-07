-- +goose Up
CREATE TABLE chirps (
    id  VARCHAR(255) NOT NULL PRIMARY KEY,
    body TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
    ON DELETE CASCADE
);


-- +goose Down
DROP TABLE chirps;