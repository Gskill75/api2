
CREATE TABLE namespaces (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    customer_id TEXT NOT NULL,
    created_by TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
CREATE TYPE kubernetes_action_type_enum AS ENUM ('create', 'delete', 'update');
CREATE TYPE kubernetes_status_enum AS ENUM ('completed', 'failed', 'error', 'success');

CREATE TABLE kubernetes_history (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    action_type kubernetes_action_type_enum NOT NULL,
    status kubernetes_status_enum NOT NULL,
    namespace_name VARCHAR(20) NOT NULL,
    username VARCHAR(255) NOT NULL,
    error_message TEXT,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    details VARCHAR(255)
);
