CREATE TABLE harbor_projects (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    customer_id TEXT NOT NULL,
    created_by TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);