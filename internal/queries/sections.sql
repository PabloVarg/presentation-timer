-- name: GetSections :many
select *
from section
where presentation = @presentation_id
order by
    case when @direction = 'ASC' and @sort_by <> '' then @sort_by end asc,
    case when @direction = 'DESC' and @sort_by <> '' then @sort_by end desc,
    id desc
limit @query_limit
offset @query_offset
;
--
-- name: GetSectionsMetadata :one
select count(*)
from section
where presentation = @presentation_id
;
--
-- name: GetSection :one
select *
from section
where id = @id
;
--
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
--
-- name: UpdateSection :execrows
UPDATE section
SET
    name = @name,
    duration = @duration,
    position = @position
WHERE
    id = @id;
--
-- name: PatchSection :execrows
UPDATE section
SET
    name = COALESCE(sqlc.narg(name), name),
    duration = COALESCE(sqlc.narg(duration), duration),
    position = COALESCE(sqlc.narg(position), position)
WHERE
    id = @id;
--
-- name: DeleteSection :execrows
delete from section
where id = @id
;
