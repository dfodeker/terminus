-- name: CreateStore :one
insert into stores (id, name, handle, address, status, default_currency, timezone, plan, created_at, updated_at)
values (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, now(), now())
RETURNING *;



-- name: DeleteAllStores :exec
DELETE FROM stores;

-- name: GetStores :many
SELECT * FROM stores ORDER BY created_at ASC;



-- name: GetStoreByHandle :one
SELECT * FROM stores WHERE handle = $1;

-- name: UpdateStore :one
UPDATE stores
SET 
    name = $2,
    handle = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;