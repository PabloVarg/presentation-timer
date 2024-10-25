-- name: CreateSection :one
INSERT INTO section (
    presentation,
    name,
    duration,
    position
) VALUES (
    @presentation,
    @name,
    @duration,
    @position
) RETURNING *;
