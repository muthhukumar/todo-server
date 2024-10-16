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
alter table tasks add column metadata VARCHAR(255) NOT NULL DEFAULT '';

alter table tasks add column start_date VARCHAR(255) NOT NULL DEFAULT '';
alter table tasks add column recurrence_pattern 
alter table tasks add column recurrence_internal INT;

CREATE TABLE url_titles (
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    title TEXT,
    is_valid BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE log (
    id SERIAL PRIMARY KEY,
    log TEXT NOT NULL,
    level VARCHAR(255) DEFAULT 'INFO',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sub_tasks (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    task_id INT NOT NULL,  -- Foreign key to reference the tasks table
    CONSTRAINT fk_task
        FOREIGN KEY(task_id) 
        REFERENCES tasks(id) 
        ON DELETE CASCADE
);


ALTER TABLE tasks 
ADD COLUMN start_date VARCHAR(255) NOT NULL DEFAULT '';

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'recurrence_pattern_enum') THEN
        CREATE TYPE recurrence_pattern_enum AS ENUM ('daily', 'weekly', 'monthly', 'yearly');
    END IF;
END $$;

ALTER TABLE tasks 
ADD COLUMN recurrence_pattern recurrence_pattern_enum;

ALTER TABLE tasks 
ADD COLUMN recurrence_interval INT;

