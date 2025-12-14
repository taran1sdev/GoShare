DROP TABLE IF EXISTS sessions;

CREATE TABLE sessions(
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    token_hash TEXT UNIQUE NOT NULL
);
