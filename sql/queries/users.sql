-- name: CreateUser :one
insert into
    users(id, created_at, udpated_at, email)
values (
    gen_random_uuid(), NOW(), NOW(), $1
)
RETURNING *;
