-- +goose Up
-- +goose StatementBegin
CREATE TABLE users( 
    id SERIAL PRIMARY KEY,
    forename TEXT,
    surname TEXT,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
