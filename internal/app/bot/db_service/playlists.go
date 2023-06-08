package db_service

import (
	"bobot/internal/app/repository"
	"context"
	"errors"
	"fmt"
)

var ErrObjectAlreadyExists = errors.New("object already exists")
var ErrObjectNotFound = errors.New("object not found")

func (s *DbService) AddUsersPL(ctx context.Context, name string, discordID string) error {
	userRow, err := s.usersRepo.GetByDiscordID(ctx, discordID)
	if err != nil {
		return fmt.Errorf("failed to find user (to add playlist): %v", err)
	}

	if _, err = s.playlistsRepo.GetByName(ctx, name); err == nil {
		return ErrObjectAlreadyExists
	} else if err != repository.ErrObjectNotFound {
		return fmt.Errorf("failed to check if playlist already exists: %v", err)
	}

	if _, err = s.playlistsRepo.Add(ctx, &repository.Playlist{
		UserID: userRow.ID,
		Name:   name,
	}); err != nil {
		return fmt.Errorf("failed to add playlist: %v", err)
	}

	return nil
}

func (s *DbService) GetPLByID(ctx context.Context, id int64) (*Playlist, error) {
	playlistRow, err := s.playlistsRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlists: %v", err)
	}

	return mapPlaylistFromRow(playlistRow), nil
}

func (s *DbService) GetUsersPLs(ctx context.Context, discordID string) ([]*Playlist, error) {
	userRow, err := s.usersRepo.GetByDiscordID(ctx, discordID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user (to get playlists): %v", err)
	}

	playlistsRows, err := s.playlistsRepo.GetByUserID(ctx, userRow.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlists: %v", err)
	}

	return mapRowsToPlaylists(playlistsRows), nil
}

func (s *DbService) DeletePLsByUser(ctx context.Context, discordID string) error {
	userRow, err := s.usersRepo.GetByDiscordID(ctx, discordID)
	if err != nil {
		return fmt.Errorf("failed to find user (to delete playlists): %v", err)
	}

	playlistsRows, err := s.playlistsRepo.GetByUserID(ctx, userRow.ID)
	if err != nil {
		return fmt.Errorf("failed to get playlists: %v", err)
	}

	for _, row := range playlistsRows {
		err = s.DeletePLByID(ctx, row.ID)
		if err != nil {
			return fmt.Errorf("failed to delete playlist id=%d : %v", row.ID, err)
		}
	}

	return nil
}

func (s *DbService) DeletePLByID(ctx context.Context, id int64) error {
	if _, err := s.playlistsRepo.GetByID(ctx, id); err == repository.ErrObjectNotFound {
		return ErrObjectNotFound
	} else if err != nil {
		return fmt.Errorf("failed to check if playlist exists: %v", err)
	}

	if err := s.DeleteTrackByPL(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tracks by playlist: %v", err)
	}

	_, err := s.playlistsRepo.DeleteByID(ctx, id)
	return err
}
