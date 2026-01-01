-- name: CreateUser :one
insert into users (id, created_at, updated_at, email)
values (gen_random_uuid(),now(), now(), $1)
RETURNING *;



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
