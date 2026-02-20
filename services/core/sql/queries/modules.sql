-- name: RegisterModule :one
INSERT INTO core.module_registry (name, version, grpc_address, health_status)
VALUES ($1, $2, $3, 'healthy')
ON CONFLICT (name) DO UPDATE SET
    version = EXCLUDED.version,
    grpc_address = EXCLUDED.grpc_address,
    health_status = 'healthy',
    last_health_check = NOW()
RETURNING *;

-- name: UnregisterModule :exec
DELETE FROM core.module_registry WHERE name = $1;

-- name: GetModuleByName :one
SELECT * FROM core.module_registry WHERE name = $1;

-- name: ListModules :many
SELECT * FROM core.module_registry ORDER BY name;

-- name: UpdateModuleHealth :exec
UPDATE core.module_registry
SET health_status = $2, last_health_check = NOW()
WHERE name = $1;
