-- name: GetPresentations :many
SELECT
    *
FROM
    presentation;
--
-- name: CreatePresentation :one
INSERT INTO presentation(
    name
) VALUES (
    @name
)
RETURNING
    *;
--
-- name: GetPresentation :one
SELECT
    *
FROM
    presentation
WHERE
    id = @id;
