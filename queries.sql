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

