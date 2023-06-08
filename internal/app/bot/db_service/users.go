package db_service

import (
	"bobot/internal/app/repository"
	"context"
	"fmt"
)

func (s *DbService) AddUser(ctx context.Context, discordID string) error {
	if _, err := s.usersRepo.GetByDiscordID(ctx, discordID); err == nil {
		return nil
	} else if err != repository.ErrObjectNotFound {
		return fmt.Errorf("failed to check if user already exists: %v", err)
	}

	if _, err := s.usersRepo.Add(ctx, &repository.User{
		DiscordID: discordID,
	}); err != nil {
		return fmt.Errorf("failed to add user: %v", err)
	}

	return nil
}

func (s *DbService) DeleteUser(ctx context.Context, discordID string) error {
	userRow, err := s.usersRepo.GetByDiscordID(ctx, discordID)
	if err != nil {
		return fmt.Errorf("failed to find user: %v", err)
	}

	if err = s.DeletePLsByUser(ctx, discordID); err != nil {
		return fmt.Errorf("failed to delete user's playlists: %v", err)
	}

	if _, err = s.usersRepo.DeleteByID(ctx, userRow.ID); err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	return nil
}
