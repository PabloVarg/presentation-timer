-- name: GetPresentations :many
select *
from presentation
;
--
-- name: GetPresentation :one
select *
from presentation
where id = @id
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
-- name: DeletePresentation :execrows
delete from presentation
where id = @id
;
