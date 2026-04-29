-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateAvatarKey :exec 
UPDATE users 
SET avatar_key = $1 
WHERE id = $2; 

-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, username, display_name, avatar_key)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
