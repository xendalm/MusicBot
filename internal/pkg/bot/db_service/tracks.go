package db_service

import (
	"bobot/internal/pkg/repository"
	"context"
	"encoding/json"
	"fmt"
)

func (s *DbService) AddTrack(ctx context.Context, track *Track) error {
	if _, err := s.playlistsRepo.GetByID(ctx, track.PlaylistID); err == repository.ErrObjectNotFound {
		return ErrObjectNotFound
	} else if err != nil {
		return fmt.Errorf("failed to find playlist (to add track): %v", err)
	}

	trackInfo, err := json.Marshal(track.Value)
	if err != nil {
		return fmt.Errorf("error while marshalling track to json: %v", err)
	}

	if _, err := s.tracksRepo.Add(ctx, &repository.Track{
		PlaylistID: track.PlaylistID,
		Info:       trackInfo,
	}); err != nil {
		return fmt.Errorf("failed to add track to playlist: %v", err)
	}

	return nil
}

func (s *DbService) GetTracksByPlaylist(ctx context.Context, plID int64) ([]*Track, error) {
	playlistRow, err := s.playlistsRepo.GetByID(ctx, plID)
	if err == repository.ErrObjectNotFound {
		return nil, ErrObjectNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to find playlist (to add track): %v", err)
	}

	tracksRows, err := s.tracksRepo.GetByPLID(ctx, playlistRow.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracks: %v", err)
	}

	tracks, err := mapRowsToTracks(tracksRows)
	if err != nil {
		return nil, fmt.Errorf("failed map rows to tracks: %v", err)
	}

	return tracks, nil
}

func (s *DbService) DeleteTrackByPL(ctx context.Context, plID int64) error {
	_, err := s.tracksRepo.DeleteByPLID(ctx, plID)
	return err
}
