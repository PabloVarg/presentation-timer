// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: sections.sql

package queries

import (
	"context"

	"time"
)

const createSection = `-- name: CreateSection :one
INSERT INTO section (
    presentation,
    name,
    duration,
    position
) VALUES (
    $1,
    $2,
    $3,
    $4
) RETURNING id, presentation, name, duration, position
`

type CreateSectionParams struct {
	Presentation int64         `json:"presentation"`
	Name         string        `json:"name"`
	Duration     time.Duration `json:"duration"`
	Position     int16         `json:"position"`
}

func (q *Queries) CreateSection(ctx context.Context, arg CreateSectionParams) (Section, error) {
	row := q.db.QueryRow(ctx, createSection,
		arg.Presentation,
		arg.Name,
		arg.Duration,
		arg.Position,
	)
	var i Section
	err := row.Scan(
		&i.ID,
		&i.Presentation,
		&i.Name,
		&i.Duration,
		&i.Position,
	)
	return i, err
}

const deleteSection = `-- name: DeleteSection :execrows
delete from section
where id = $1
`

func (q *Queries) DeleteSection(ctx context.Context, id int64) (int64, error) {
	result, err := q.db.Exec(ctx, deleteSection, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getSection = `-- name: GetSection :one
select id, presentation, name, duration, position
from section
where id = $1
`

func (q *Queries) GetSection(ctx context.Context, id int64) (Section, error) {
	row := q.db.QueryRow(ctx, getSection, id)
	var i Section
	err := row.Scan(
		&i.ID,
		&i.Presentation,
		&i.Name,
		&i.Duration,
		&i.Position,
	)
	return i, err
}

const getSections = `-- name: GetSections :many
select id, presentation, name, duration, position
from section
where presentation = $1
order by
    case when $2 = 'ASC' and $3 <> '' then $3 end asc,
    case when $2 = 'DESC' and $3 <> '' then $3 end desc,
    id desc
limit $5
offset $4
`

type GetSectionsParams struct {
	PresentationID int64       `json:"presentation_id"`
	Direction      interface{} `json:"direction"`
	SortBy         interface{} `json:"sort_by"`
	QueryOffset    int32       `json:"query_offset"`
	QueryLimit     int32       `json:"query_limit"`
}

func (q *Queries) GetSections(ctx context.Context, arg GetSectionsParams) ([]Section, error) {
	rows, err := q.db.Query(ctx, getSections,
		arg.PresentationID,
		arg.Direction,
		arg.SortBy,
		arg.QueryOffset,
		arg.QueryLimit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Section
	for rows.Next() {
		var i Section
		if err := rows.Scan(
			&i.ID,
			&i.Presentation,
			&i.Name,
			&i.Duration,
			&i.Position,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSectionsMetadata = `-- name: GetSectionsMetadata :one
select count(*)
from section
where presentation = $1
`

func (q *Queries) GetSectionsMetadata(ctx context.Context, presentationID int64) (int64, error) {
	row := q.db.QueryRow(ctx, getSectionsMetadata, presentationID)
	var count int64
	err := row.Scan(&count)
	return count, err
}
