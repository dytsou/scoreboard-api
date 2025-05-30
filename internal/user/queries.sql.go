// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: queries.sql

package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const create = `-- name: Create :one
INSERT INTO users (
    id, email, name, given_name, family_name, picture, email_verified, locale, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    $1, $2, $3, $4, $5, $6, $7,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
RETURNING id, email, name, given_name, family_name, picture, email_verified, locale, created_at, updated_at
`

type CreateParams struct {
	Email         string
	Name          pgtype.Text
	GivenName     pgtype.Text
	FamilyName    pgtype.Text
	Picture       pgtype.Text
	EmailVerified bool
	Locale        pgtype.Text
}

func (q *Queries) Create(ctx context.Context, arg CreateParams) (User, error) {
	row := q.db.QueryRow(ctx, create,
		arg.Email,
		arg.Name,
		arg.GivenName,
		arg.FamilyName,
		arg.Picture,
		arg.EmailVerified,
		arg.Locale,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.GivenName,
		&i.FamilyName,
		&i.Picture,
		&i.EmailVerified,
		&i.Locale,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const delete = `-- name: Delete :exec
DELETE FROM users WHERE id = $1
`

func (q *Queries) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, delete, id)
	return err
}

const existsByEmail = `-- name: ExistsByEmail :one
SELECT EXISTS(
    SELECT 1 FROM users WHERE email = $1
) AS email_exists
`

func (q *Queries) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	row := q.db.QueryRow(ctx, existsByEmail, email)
	var email_exists bool
	err := row.Scan(&email_exists)
	return email_exists, err
}

const getByEmail = `-- name: GetByEmail :one
SELECT id, email, name, given_name, family_name, picture, email_verified, locale, created_at, updated_at FROM users WHERE email = $1
`

func (q *Queries) GetByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.GivenName,
		&i.FamilyName,
		&i.Picture,
		&i.EmailVerified,
		&i.Locale,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getByID = `-- name: GetByID :one
SELECT id, email, name, given_name, family_name, picture, email_verified, locale, created_at, updated_at FROM users WHERE id = $1
`

func (q *Queries) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRow(ctx, getByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.GivenName,
		&i.FamilyName,
		&i.Picture,
		&i.EmailVerified,
		&i.Locale,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const update = `-- name: Update :one
UPDATE users SET 
    name = $2,
    given_name = $3,
    family_name = $4,
    picture = $5,
    email_verified = $6,
    locale = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE email = $1
RETURNING id, email, name, given_name, family_name, picture, email_verified, locale, created_at, updated_at
`

type UpdateParams struct {
	Email         string
	Name          pgtype.Text
	GivenName     pgtype.Text
	FamilyName    pgtype.Text
	Picture       pgtype.Text
	EmailVerified bool
	Locale        pgtype.Text
}

func (q *Queries) Update(ctx context.Context, arg UpdateParams) (User, error) {
	row := q.db.QueryRow(ctx, update,
		arg.Email,
		arg.Name,
		arg.GivenName,
		arg.FamilyName,
		arg.Picture,
		arg.EmailVerified,
		arg.Locale,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.GivenName,
		&i.FamilyName,
		&i.Picture,
		&i.EmailVerified,
		&i.Locale,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
