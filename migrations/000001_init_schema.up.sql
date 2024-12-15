CREATE TABLE IF NOT EXISTS users (
    id integer NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    username varchar(50) UNIQUE,
    password varchar(50)
);

CREATE INDEX user_id ON users USING btree (id);