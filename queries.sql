CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT false,
    completed_on VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    marked_today VARCHAR(255) DEFAULT '',
    is_important BOOLEAN DEFAULT false
);

alter table tasks add column due_date VARCHAR(255) NOT NULL DEFAULT '';