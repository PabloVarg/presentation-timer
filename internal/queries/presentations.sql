-- name: GetPresentations :many
select *
from presentation
order by
    case when @direction::text = 'ASC' and @sort_by::text <> '' then @sort_by end asc,
    case when @direction::text = 'DESC' and @sort_by::text <> '' then @sort_by end desc,
    id desc
limit @query_limit
offset @query_offset
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
