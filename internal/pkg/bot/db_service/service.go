package db_service

import (
	"bobot/internal/pkg/db"
	"bobot/internal/pkg/repository"
	"bobot/internal/pkg/repository/postgresql"
)

type DbService struct {
	usersRepo     repository.UsersRepo
	playlistsRepo repository.PlaylistsRepo
	tracksRepo    repository.TracksRepo
}

func NewDbService(db db.DB) *DbService {
	return &DbService{
		usersRepo:     postgresql.NewUsersRepo(db),
		playlistsRepo: postgresql.NewPLRepo(db),
		tracksRepo:    postgresql.NewTracksRepo(db),
	}
}
