-- name: CreateUser :one
INSERT INTO users (id, gid, created_at, updated_at, email, hashed_password)
VALUES (gen_random_uuid(), $1, now(), now(), $2, $3)
RETURNING *;

-- name: GetUserByGID :one
SELECT * FROM users
WHERE gid = $1;



-- name: DeleteAllUsers :exec
DELETE FROM users;


-- name: GetAllUsers :many
SELECT * FROM users ORDER BY created_at ASC;



-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;


-- name: UpdateUser :one
UPDATE users
SET 
    email = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;
