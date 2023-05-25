package db_service

import (
	"bobot/internal/pkg/repository"
	"encoding/json"
	"fmt"
	"github.com/disgoorg/disgolink/v2/lavalink"
)

type Playlist struct {
	ID     int64
	UserID int64
	Name   string
}

type Track struct {
	ID         int64
	PlaylistID int64
	Value      *lavalink.Track
}

func mapPlaylistFromRow(playlistsRow *repository.Playlist) *Playlist {
	return &Playlist{
		ID:     playlistsRow.ID,
		UserID: playlistsRow.UserID,
		Name:   playlistsRow.Name,
	}
}

func mapRowsToPlaylists(playlistsRows []*repository.Playlist) []*Playlist {
	playlists := make([]*Playlist, len(playlistsRows))
	for i, playlistRow := range playlistsRows {
		playlists[i] = mapPlaylistFromRow(playlistRow)
	}

	return playlists
}

func mapTrackFromRow(trackRow *repository.Track) (*Track, error) {
	var track lavalink.Track
	err := json.Unmarshal(trackRow.Info, &track)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling json to track: %v", err)
	}

	return &Track{
		ID:         trackRow.ID,
		PlaylistID: trackRow.PlaylistID,
		Value:      &track,
	}, nil
}

func mapRowsToTracks(tracksRows []*repository.Track) ([]*Track, error) {
	tracks := make([]*Track, len(tracksRows))
	for i, trackRow := range tracksRows {
		track, err := mapTrackFromRow(trackRow)
		if err != nil {
			return nil, err
		}
		tracks[i] = track
	}

	return tracks, nil
}
