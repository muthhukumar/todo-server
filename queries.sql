CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT false,
    completed_on VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    marked_today VARCHAR(255) NOT NULL DEFAULT ''
);

ALTER TABLE tasks
ADD COLUMN marked_today VARCHAR(255) NOT NULL DEFAULT '';

alter table tasks add column is_important BOOL default false;