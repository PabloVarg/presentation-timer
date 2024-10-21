// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: presentation.sql

package queries

import (
	"context"
)

const createPresentation = `-- name: CreatePresentation :one
INSERT INTO presentation(
    name
) VALUES (
    $1
)
RETURNING id, name
`

func (q *Queries) CreatePresentation(ctx context.Context, name string) (Presentation, error) {
	row := q.db.QueryRow(ctx, createPresentation, name)
	var i Presentation
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const getPresentation = `-- name: GetPresentation :one
SELECT
    id, name
FROM
    presentation
WHERE
    id = $1
`

func (q *Queries) GetPresentation(ctx context.Context, id int64) (Presentation, error) {
	row := q.db.QueryRow(ctx, getPresentation, id)
	var i Presentation
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const getPresentations = `-- name: GetPresentations :many
SELECT
    id, name
FROM
    presentation
`

func (q *Queries) GetPresentations(ctx context.Context) ([]Presentation, error) {
	rows, err := q.db.Query(ctx, getPresentations)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Presentation
	for rows.Next() {
		var i Presentation
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
