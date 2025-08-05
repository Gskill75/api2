-- name: InsertHarborProject :exec
INSERT INTO harbor_projects (
    name,
    customer_id,
    created_by
) VALUES (
    $1, $2, $3
);
