-- name: GetPresentations :many
select presentation.*, coalesce(sum(section.duration), '0 seconds')::interval duration
from presentation
left join section on presentation.id = section.presentation
group by presentation.id
order by
    case
        when @direction::text = 'ASC' and @sort_by::text = 'name' then presentation.name
    end asc,
    case
        when @direction::text = 'DESC' and @sort_by::text = 'name'
        then presentation.name
    end desc,
    presentation.id desc
limit @query_limit
offset @query_offset
;
--
-- name: GetPresentationsMetadata :one
select count(*)
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
-- name: UpdatePresentation :execrows
UPDATE presentation
SET name = @name
WHERE id = @id;
--
-- name: PatchPresentation :execrows
UPDATE presentation
SET name = COALESCE(sqlc.narg('name'), name)
WHERE id = @id;
--
-- name: DeletePresentation :execrows
delete from presentation
where id = @id
;
