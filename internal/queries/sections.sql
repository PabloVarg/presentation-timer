-- name: GetSections :many
select *
from section
where presentation = @presentation_id
order by
    case when @direction::text = 'ASC' and @sort_by::text = 'name' then name end asc,
    case when @direction::text = 'DESC' and @sort_by::text = 'name' then name end desc,
    case
        when @direction::text = 'ASC' and @sort_by::text = 'duration' then duration
    end asc,
    case
        when @direction::text = 'DESC' and @sort_by::text = 'duration' then duration
    end desc,
    case
        when @direction::text = 'ASC' and @sort_by::text = 'position' then position
    end asc,
    case
        when @direction::text = 'DESC' and @sort_by::text = 'position' then position
    end desc,
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
--
-- name: MaxPosition :one
select coalesce(max(position), 0)::smallint
from section
where presentation = @presentation_id
;
--
-- name: CleanPositions :exec
call clean_section_positions()
;
--
-- name: CleanPositionsBySectionGroup :exec
with
    ordered as (
        select id, row_number() over (order by position) as new_position
        from section o
        where o.presentation = (select i.presentation from section i where i.id = @id)
    )
    update section
    set position = ordered.new_position
from ordered
where section.id = ordered.id
;
--
-- name: MoveSection :exec
update section s
set position = case when s.id <> $1 then position - ($2::int / abs($2)) when id = $1 then position + $2 end
where position between least(position + $2, position) and greatest(position + $2, position) and presentation = (
    select sp.presentation from section sp where sp.id = $1
)
;
--
-- name: GetSectionsByPosition :many
with
    ordered as (
        select s.id, row_number() over (order by s.position) as new_position
        from section s
        where s.presentation = @presentation_id
    )
select o.*
from section o
inner join ordered ord on ord.id = o.id
where o.presentation = @presentation_id
order by ord.new_position
;
