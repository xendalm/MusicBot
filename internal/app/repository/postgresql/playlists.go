package postgresql

import (
	"bobot/internal/app/db"
	"bobot/internal/app/repository"
	"context"
	"database/sql"
	"errors"
)

type PLRepo struct {
	db db.DB
}

func NewPLRepo(db db.DB) *PLRepo {
	return &PLRepo{db: db}
}

func (r *PLRepo) Add(ctx context.Context, playlist *repository.Playlist) (int64, error) {
	var id int64
	err := r.db.ExecQueryRow(ctx, `INSERT INTO playlists(user_id, name) VALUES ($1, $2) RETURNING id`,
		playlist.UserID, playlist.Name).Scan(&id)
	return id, err
}

func (r *PLRepo) GetByID(ctx context.Context, id int64) (*repository.Playlist, error) {
	var playlist repository.Playlist
	err := r.db.Get(ctx, &playlist, "SELECT id,user_id,name,created_at,updated_at FROM playlists WHERE id = $1", id)
	if errors.As(err, &sql.ErrNoRows) {
		return nil, repository.ErrObjectNotFound
	}
	return &playlist, err
}

func (r *PLRepo) GetByName(ctx context.Context, name string) (*repository.Playlist, error) {
	var playlist repository.Playlist
	err := r.db.Get(ctx, &playlist, "SELECT id,user_id,name,created_at,updated_at FROM playlists WHERE name = $1", name)
	if errors.As(err, &sql.ErrNoRows) {
		return nil, repository.ErrObjectNotFound
	}
	return &playlist, err
}

func (r *PLRepo) GetByUserID(ctx context.Context, userID int64) ([]*repository.Playlist, error) {
	playlists := make([]*repository.Playlist, 0)
	err := r.db.Select(ctx, &playlists, "SELECT id,user_id,name,created_at,updated_at FROM playlists WHERE user_id = $1", userID)
	return playlists, err
}

func (r *PLRepo) Update(ctx context.Context, playlist *repository.Playlist) (bool, error) {
	result, err := r.db.Exec(ctx, "UPDATE playlists SET user_id = $1, name = $2, updated_at = $3 WHERE id = $4",
		playlist.UserID, playlist.Name, playlist.UpdatedAt, playlist.ID)
	return result.RowsAffected() > 0, err
}

func (r *PLRepo) DeleteByID(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.Exec(ctx,
		"DELETE FROM playlists WHERE id = $1", id)
	return result.RowsAffected() > 0, err
}
