-- name: GetNamespace :one
SELECT * FROM namespaces WHERE name = $1 AND customer_id = $2;

-- name: GetNamespaceByCustomer :many
SELECT * FROM namespaces WHERE customer_id = $1;
-- name: InsertNamespace :exec
INSERT INTO namespaces (
    name,
    customer_id,
    created_by
) VALUES (
    $1, $2, $3
);

-- name: ListNamespacesByCustomerID :many
SELECT * FROM namespaces WHERE customer_id = $1 ORDER BY created_at DESC;

-- name: DeleteNamespace :one
DELETE FROM namespaces
WHERE name = $1 AND customer_id = $2
RETURNING name, customer_id;

-- name: GetHistoryByCustomer :many
SELECT * FROM kubernetes_history
WHERE customer_id = $1
ORDER BY created_at DESC;

-- name: CreateHistory :one
INSERT INTO kubernetes_history (
  customer_id, action_type, status, namespace_name, username, created_by, details, error_message
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;
