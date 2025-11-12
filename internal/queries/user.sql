-- Users Table -- 
-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC;

-- name: ListActiveUsers :many
SELECT * FROM users
WHERE is_active = true
ORDER BY created_at DESC;

-- name: ListUsersByRole :many
SELECT * FROM users
WHERE role = $1
ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (
  email,
  password_hash,
  username,
  login_method,
  first_name,
  last_name,
  profile_picture_url,
  timezone,
  is_active,
  email_verified,
  role
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users
SET
  email = $2,
  username = $3,
  first_name = $4,
  last_name = $5,
  profile_picture_url = $6,
  timezone = $7,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET
  password_hash = $2,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateUserRole :exec
UPDATE users
SET
  role = $2,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateUserLoginMethod :exec
UPDATE users
SET
  login_method = $2,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: VerifyUserEmail :exec
UPDATE users
SET
  email_verified = true,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users
SET
  is_active = false,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: ActivateUser :exec
UPDATE users
SET
  is_active = true,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateLastLogin :exec
UPDATE users
SET
  last_login_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CountUsersByRole :one
SELECT COUNT(*) FROM users
WHERE role = $1;

-- name: SearchUsersByEmail :many
SELECT * FROM users
WHERE email ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: SearchUsersByName :many
SELECT * FROM users
WHERE
  first_name ILIKE '%' || $1 || '%' OR
  last_name ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;