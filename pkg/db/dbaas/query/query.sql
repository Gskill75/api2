-- name: GetHistory :one
SELECT * FROM awx_history
WHERE id = $1 LIMIT 1;

-- name: GetHistoryByJobID :one
SELECT * FROM awx_history
WHERE awx_job_id = $1 LIMIT 1;

-- name: GetHistoryByCustomer :many
SELECT * FROM awx_history
WHERE customer_id = $1
ORDER BY created_at DESC;

-- name: GetHistoryByStatus :many
SELECT * FROM awx_history
WHERE status = $1
ORDER BY created_at DESC;

-- name: GetActiveJobs :many
SELECT * FROM awx_history
WHERE status IN ('pending', 'running')
ORDER BY created_at DESC;

-- name: CreateHistory :one
INSERT INTO awx_history (
  customer_id, awx_job_id, awx_template_name, awx_template_id, 
  action_type, status, instance_name, username, extra_vars, 
  awx_status, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: UpdateHistoryStatus :exec
UPDATE awx_history 
SET status = $2, awx_status = $3, updated_at = NOW()
WHERE id = $1;

-- name: UpdateHistoryCompletion :exec
UPDATE awx_history 
SET status = $2, awx_status = $3, completed_at = NOW(), 
    error_message = $4, updated_at = NOW()
WHERE id = $1;

-- name: UpdateHistoryByJobID :exec
UPDATE awx_history 
SET status = $2, awx_status = $3, updated_at = NOW()
WHERE awx_job_id = $1;

-- name: DeleteHistory :exec
DELETE FROM awx_history WHERE id = $1;

-- Database Instances Queries

-- name: GetDBInstance :one
SELECT * FROM db_instances
WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: GetDBInstanceByName :one
SELECT * FROM db_instances
WHERE instance_name = $1 AND customer_id = $2 AND deleted_at IS NULL LIMIT 1;

-- name: GetDBInstancesByCustomer :many
SELECT * FROM db_instances
WHERE customer_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetDBInstancesByStatus :many
SELECT * FROM db_instances
WHERE status = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateDBInstance :one
INSERT INTO db_instances (
  customer_id, db_type, version, host, port, username, 
  status, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateDBInstanceStatus :exec
UPDATE db_instances 
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateDBInstanceDetails :exec
UPDATE db_instances 
SET host = $2, port = $3, version = $4, updated_at = NOW()
WHERE id = $1;

-- name: SoftDeleteDBInstance :exec
UPDATE db_instances 
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1;