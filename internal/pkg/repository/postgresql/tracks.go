package postgresql

import (
	"bobot/internal/pkg/db"
	"bobot/internal/pkg/repository"
	"context"
	"database/sql"
	"errors"
)

type TracksRepo struct {
	db db.DB
}

func NewTracksRepo(db db.DB) *TracksRepo {
	return &TracksRepo{db: db}
}

func (r *TracksRepo) Add(ctx context.Context, track *repository.Track) (int64, error) {
	var id int64
	err := r.db.ExecQueryRow(ctx, `INSERT INTO tracks(playlist_id, info) VALUES ($1, $2) RETURNING id`,
		track.PlaylistID, track.Info).Scan(&id)
	return id, err
}

func (r *TracksRepo) GetByID(ctx context.Context, id int64) (*repository.Track, error) {
	var track repository.Track
	err := r.db.Get(ctx, &track, "SELECT id,playlist_id,info,created_at,updated_at FROM tracks WHERE id = $1", id)
	if errors.As(err, &sql.ErrNoRows) {
		return nil, repository.ErrObjectNotFound
	}
	return &track, err
}

func (r *TracksRepo) GetByPLID(ctx context.Context, plID int64) ([]*repository.Track, error) {
	tracks := make([]*repository.Track, 0)
	err := r.db.Select(ctx, &tracks, "SELECT id,playlist_id,info,created_at,updated_at FROM tracks WHERE playlist_id = $1", plID)
	return tracks, err
}

func (r *TracksRepo) Update(ctx context.Context, track *repository.Track) (bool, error) {
	result, err := r.db.Exec(ctx, "UPDATE tracks SET playlist_id = $1, info = $2, updated_at = $3 WHERE id = $4",
		track.PlaylistID, track.Info, track.UpdatedAt, track.ID)
	return result.RowsAffected() > 0, err
}

func (r *TracksRepo) DeleteByID(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.Exec(ctx,
		"DELETE FROM tracks WHERE id = $1", id)
	return result.RowsAffected() > 0, err
}

func (r *TracksRepo) DeleteByPLID(ctx context.Context, plID int64) (bool, error) {
	result, err := r.db.Exec(ctx,
		"DELETE FROM tracks WHERE playlist_id = $1", plID)
	return result.RowsAffected() > 0, err
}
