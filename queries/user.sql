-- name: CreateUser :one
INSERT INTO "user" (username, password_hash)
VALUES ($1, $2)
RETURNING id, username, created_at, updated_at;

-- name: GetUserByName :one
SELECT id, username, password_hash, created_at, updated_at
FROM "user"
WHERE username = $1;

-- name: GetAccessTokenByHash :one
SELECT id, user_id, token_hash, expires_at, created_at
FROM access_token
WHERE token_hash = $1;

-- name: GetRoleByUserID :one
SELECT r.id, r.name, r.description
FROM role r
JOIN user_role ur ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: GetRoleByAccessTokenID :one
SELECT r.id, r.name, r.description
FROM role r
JOIN access_token_role atr ON r.id = atr.role_id
WHERE atr.access_token_id = $1;

-- name: CreateAuditLog :one
INSERT INTO audit_log (user_id, action, details)
VALUES ($1, $2, $3)
RETURNING id, user_id, action, details, created_at;

-- name: UpdateUserRole :exec
INSERT INTO user_role (user_id, role_id)
VALUES ($1, (SELECT id FROM role WHERE name = $2))
ON CONFLICT (user_id) DO UPDATE SET role_id = EXCLUDED.role_id;

-- name: UpdateAccessTokenRole :exec
INSERT INTO access_token_role (access_token_id, role_id)
VALUES ($1, (SELECT id FROM role WHERE name = $2))
ON CONFLICT (access_token_id) DO UPDATE SET role_id = EXCLUDED.role_id;

-- name: DeleteAccessToken :exec
DELETE FROM access_token
WHERE id = $1;
