DROP TABLE IF EXISTS users;

CREATE TABLE users( 
    id SERIAL PRIMARY KEY,
    forename TEXT,
    surname TEXT,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);
