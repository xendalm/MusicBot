package repository

import (
	"context"
	"errors"
)

var ErrObjectNotFound = errors.New("object not found")

type UsersRepo interface {
	Add(ctx context.Context, user *User) (int64, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByDiscordID(ctx context.Context, discordID string) (*User, error)
	Update(ctx context.Context, user *User) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
}

type PlaylistsRepo interface {
	Add(ctx context.Context, playlist *Playlist) (int64, error)
	GetByID(ctx context.Context, id int64) (*Playlist, error)
	GetByName(ctx context.Context, name string) (*Playlist, error)
	GetByUserID(ctx context.Context, userID int64) ([]*Playlist, error)
	Update(ctx context.Context, playlist *Playlist) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
}

type TracksRepo interface {
	Add(ctx context.Context, track *Track) (int64, error)
	GetByID(ctx context.Context, id int64) (*Track, error)
	GetByPLID(ctx context.Context, plID int64) ([]*Track, error)
	Update(ctx context.Context, track *Track) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
	DeleteByPLID(ctx context.Context, plID int64) (bool, error)
}
