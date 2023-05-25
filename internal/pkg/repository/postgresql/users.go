package postgresql

import (
	"bobot/internal/pkg/db"
	"bobot/internal/pkg/repository"
	"context"
	"database/sql"
	"errors"
)

type UsersRepo struct {
	db db.DB
}

func NewUsersRepo(db db.DB) *UsersRepo {
	return &UsersRepo{db: db}
}

func (r *UsersRepo) Add(ctx context.Context, user *repository.User) (int64, error) {
	var id int64
	err := r.db.ExecQueryRow(ctx, `INSERT INTO users(discord_id) VALUES ($1) RETURNING id`, user.DiscordID).Scan(&id)
	return id, err
}

func (r *UsersRepo) GetByID(ctx context.Context, id int64) (*repository.User, error) {
	var user repository.User
	err := r.db.Get(ctx, &user, "SELECT id,discord_id,created_at,updated_at FROM users WHERE id = $1", id)
	if errors.As(err, &sql.ErrNoRows) {
		return nil, repository.ErrObjectNotFound
	}
	return &user, err
}

func (r *UsersRepo) GetByDiscordID(ctx context.Context, discordID string) (*repository.User, error) {
	var user repository.User
	err := r.db.Get(ctx, &user, "SELECT id,discord_id,created_at,updated_at FROM users WHERE discord_id = $1", discordID)
	if errors.As(err, &sql.ErrNoRows) {
		return nil, repository.ErrObjectNotFound
	}
	return &user, err
}

func (r *UsersRepo) Update(ctx context.Context, user *repository.User) (bool, error) {
	result, err := r.db.Exec(ctx, "UPDATE users SET discord_id = $1, updated_at = $2 WHERE id = $3",
		user.DiscordID, user.UpdatedAt, user.ID)
	return result.RowsAffected() > 0, err
}

func (r *UsersRepo) DeleteByID(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.Exec(ctx,
		"DELETE FROM users WHERE id = $1", id)
	return result.RowsAffected() > 0, err
}
