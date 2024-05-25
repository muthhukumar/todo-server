CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT false,
    completed_on DATE
);

ALTER TABLE tasks ALTER COLUMN completed_on TYPE TEXT;