package repository

import "time"

type User struct {
	ID        int64     `db:"id"`
	DiscordID string    `db:"discord_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Playlist struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Track struct {
	ID         int64     `db:"id"`
	PlaylistID int64     `db:"playlist_id"`
	Info       []byte    `db:"info"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
