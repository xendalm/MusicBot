-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
       id bigserial primary key,
       discord_id text not null,
       created_at timestamp default now() not null,
       updated_at timestamp default now() not null
);

CREATE TABLE playlists (
       id bigserial primary key,
       user_id bigint not null,
       name text not null,
       created_at timestamp default now() not null,
       updated_at timestamp default now() not null
);

CREATE TABLE tracks (
        id bigserial primary key,
        playlist_id bigint not null,
        info jsonb NOT NULL,
        created_at timestamp default now() not null,
        updated_at timestamp default now() not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tracks;
DROP TABLE playlists;
DROP TABLE users;
-- +goose StatementEnd
