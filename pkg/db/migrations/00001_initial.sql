-- +goose Up
CREATE TYPE action_type_enum AS ENUM ('create', 'delete', 'update');
CREATE TYPE status_enum AS ENUM ('pending', 'running', 'completed', 'failed', 'canceled','error');
CREATE TYPE awx_status_enum AS ENUM ('pending', 'waiting', 'running', 'successful', 'failed', 'error', 'canceled');
 
CREATE TABLE awx_history (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    awx_job_id BIGINT,
    awx_template_name VARCHAR(255),
    awx_template_id INTEGER,
    action_type action_type_enum NOT NULL,
    status status_enum NOT NULL,
    instance_name VARCHAR(20) NOT NULL,
    username VARCHAR(255),
    extra_vars JSONB,
    awx_status awx_status_enum,
    error_message TEXT,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
 
CREATE INDEX idx_awx_history_customer_id ON awx_history(customer_id);
CREATE INDEX idx_awx_history_awx_job_id ON awx_history(awx_job_id);
CREATE INDEX idx_awx_history_status ON awx_history(status);
CREATE INDEX idx_awx_history_action_type ON awx_history(action_type);
CREATE INDEX idx_awx_history_awx_template_name ON awx_history(awx_template_name);
CREATE INDEX idx_awx_history_instance_name ON awx_history(instance_name);
 
 
  CREATE TABLE db_instances (
      id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
      customer_id VARCHAR(255) NOT NULL,
      db_type VARCHAR(50) NOT NULL,
      version VARCHAR(50),
      host VARCHAR(255),
      port INTEGER,
      username VARCHAR(255),
      status status_enum NOT NULL,
      instance_name VARCHAR(20) NOT NULL,
      created_by VARCHAR(255) NOT NULL,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      deleted_at TIMESTAMPTZ
  );
 
CREATE INDEX idx_db_instances_customer_id ON db_instances(customer_id);
CREATE INDEX idx_db_instances_status ON db_instances(status);
CREATE INDEX idx_db_instances_instance_name ON db_instances(instance_name);
 
CREATE TABLE harbor_projects (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    customer_id TEXT NOT NULL,
    created_by TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
 
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

 
-- +goose Down
DROP TABLE awx_history;
DROP TABLE db_instances;
DROP TABLE harbor_projects;
DROP TABLE namespaces;
DROP TYPE action_type_enum;
DROP TYPE status_enum;
DROP TYPE awx_status_enum;
DROP TABLE dbaas_offers;
 