-- name: GetPresentations :many
select *
from presentation
;
--
-- name: CreatePresentation :one
INSERT INTO presentation(
    name
) VALUES (
    @name
)
RETURNING *;
--
-- name: GetPresentation :one
select *
from presentation
where id = @id
;
--
-- name: DeletePresentation :execrows
delete from presentation
where id = @id
;
