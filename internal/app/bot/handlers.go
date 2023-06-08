package bot

import (
	"bobot/internal/app/bot/db_service"
	"bobot/internal/app/bot/tools"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"strconv"
	"time"
)

func ping(event *discordgo.InteractionCreate, bot *Bot) error {
	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong! ping: " + strconv.Itoa(int(bot.Session.HeartbeatLatency().Milliseconds())) + "ms.",
		},
	})
}

func play(event *discordgo.InteractionCreate, bot *Bot) error {
	bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	voiceState, err := bot.Session.State.VoiceState(event.GuildID, event.Member.User.ID)
	if err != nil {
		bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
			Embeds: tools.ErrorMsg("You are not in a voice channel. Please join one and try again"),
		})
		return err
	}

	query := event.ApplicationCommandData().Options[0].StringValue()
	if !urlPattern.MatchString(query) && !searchPattern.MatchString(query) {
		query = lavalink.SearchTypeYoutube.Apply(query)
	}

	var player disgolink.Player
	player = bot.Lavalink.ExistingPlayer(snowflake.MustParse(event.GuildID))
	if player != nil {
		curTrack := player.Track()
		if curTrack != nil && player.Paused() {
			player.Update(context.TODO(), lavalink.WithPaused(false))
		}
	} else {
		player = bot.Lavalink.Player(snowflake.MustParse(event.GuildID))
	}

	queue := bot.Queues.Get(event.GuildID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var toPlay *lavalink.Track
	bot.Lavalink.BestNode().LoadTracksHandler(ctx, query, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			if player.Track() == nil {
				toPlay = &track
				bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
					Embeds: tools.SuccessMsg(fmt.Sprintf("Playing track: [`%s`](<%s>)", track.Info.Title, *track.Info.URI)),
				})
			} else {
				queue.Add(track)
				bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
					Embeds: tools.SuccessMsg(fmt.Sprintf("Added track to queue: [`%s`](<%s>)", track.Info.Title, *track.Info.URI)),
				})
			}
		},
		func(playlist lavalink.Playlist) {
			bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
				Embeds: tools.SuccessMsg(fmt.Sprintf("Loaded playlist: `%s` with `%d` tracks", playlist.Info.Name, len(playlist.Tracks))),
			})
			if player.Track() == nil {
				toPlay = &playlist.Tracks[0]
				queue.Add(playlist.Tracks[1:]...)
			} else {
				queue.Add(playlist.Tracks...)
			}
		},
		func(tracks []lavalink.Track) {
			if player.Track() == nil {
				toPlay = &tracks[0]
				bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
					Embeds: tools.SuccessMsg(fmt.Sprintf("Playing track: [`%s`](<%s>)", tracks[0].Info.Title, *tracks[0].Info.URI)),
				})
			} else {
				queue.Add(tracks[0])
				bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
					Embeds: tools.SuccessMsg(fmt.Sprintf("Added track to queue: [`%s`](<%s>)", tracks[0].Info.Title, *tracks[0].Info.URI)),
				})
			}
		},
		func() {
			bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
				Embeds: tools.ErrorMsg(fmt.Sprintf("Nothing found for: `%s`", query)),
			})
		},
		func(err error) {
			bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
				Embeds: tools.ErrorMsg(fmt.Sprintf("Error while looking up query: `%s`", err)),
			})
		},
	))

	if toPlay == nil {
		return nil
	}

	if err = bot.Session.ChannelVoiceJoinManual(event.GuildID, voiceState.ChannelID, false, true); err != nil {
		bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
			Embeds: tools.ErrorMsg("Couldn't join the voice channel"),
		})
		return err
	}

	return player.Update(context.TODO(), lavalink.WithTrack(*toPlay))
}

func pause(event *discordgo.InteractionCreate, bot *Bot) error {
	player := bot.Lavalink.ExistingPlayer(snowflake.MustParse(event.GuildID))
	if player == nil || player.Track() == nil {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("Nothing is being played at the moment"),
			},
		})
	}

	if err := player.Update(context.TODO(), lavalink.WithPaused(!player.Paused())); err != nil {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg(fmt.Sprintf("Error while pausing/resuming: `%s`", err)),
			},
		})
	}

	if player.Paused() {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.SuccessMsg("Paused"),
			},
		})
	} else {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.SuccessMsg(fmt.Sprintf("Resumed: [`%s`](<%s>)", player.Track().Info.Title, *player.Track().Info.URI)),
			},
		})
	}
}

func stop(event *discordgo.InteractionCreate, bot *Bot) error {
	player := bot.Lavalink.ExistingPlayer(snowflake.MustParse(event.GuildID))
	if player == nil {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("Nothing is being played at the moment"),
			},
		})
	}

	if err := bot.Session.ChannelVoiceJoinManual(event.GuildID, "", false, false); err != nil {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg(fmt.Sprintf("Error while disconnecting: `%s`", err)),
			},
		})
	}

	player.Update(context.TODO(), lavalink.WithNullTrack())

	queue := bot.Queues.Get(event.GuildID)
	queue.Clear()

	bot.Lavalink.RemovePlayer(snowflake.MustParse(event.GuildID))

	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: tools.SuccessMsg("Player stopped"),
		},
	})
}

func nowPlaying(event *discordgo.InteractionCreate, bot *Bot) error {
	player := bot.Lavalink.ExistingPlayer(snowflake.MustParse(event.GuildID))
	if player == nil || player.Track() == nil {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("Nothing is being played at the moment"),
			},
		})
	}

	track := player.Track()
	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: tools.SuccessMsg(fmt.Sprintf("Now playing: [`%s`](<%s>)\n\n %s / %s", track.Info.Title, *track.Info.URI, tools.FormatPosition(player.Position()), tools.FormatPosition(track.Info.Length))),
		},
	})
}

func showQueue(event *discordgo.InteractionCreate, bot *Bot) error {
	queue := bot.Queues.Get(event.GuildID)
	if queue == nil {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("There is no queue"),
			},
		})
	}

	if len(queue.Tracks) == 0 {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.SuccessMsg("No tracks in queue"),
			},
		})
	}

	var tracksStr string
	for i, track := range queue.Tracks {
		tracksStr += fmt.Sprintf("%d. [`%s`](<%s>)\n", i+1, track.Info.Title, *track.Info.URI)
	}

	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: tools.SuccessMsg(fmt.Sprintf("Queue:\n%s", tracksStr)),
		},
	})
}

func clearQueue(event *discordgo.InteractionCreate, bot *Bot) error {
	queue := bot.Queues.Get(event.GuildID)
	if queue == nil {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("There is no queue"),
			},
		})
	}

	queue.Clear()
	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: tools.SuccessMsg("Queue cleared"),
		},
	})
}

func shuffle(event *discordgo.InteractionCreate, bot *Bot) error {
	queue := bot.Queues.Get(event.GuildID)
	if queue == nil {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("Not connected to a voice channel"),
			},
		})
	}

	if len(queue.Tracks) == 0 {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("No tracks in queue"),
			},
		})
	}

	queue.Shuffle()
	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: tools.SuccessMsg("Queue shuffled"),
		},
	})
}

func skip(event *discordgo.InteractionCreate, bot *Bot) error {
	player := bot.Lavalink.ExistingPlayer(snowflake.MustParse(event.GuildID))
	if player == nil || player.Track() == nil {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("Nothing is being played at the moment"),
			},
		})
	}

	queue := bot.Queues.Get(event.GuildID)
	nextTrack, ok := queue.Next()

	if ok {
		if err := player.Update(context.TODO(), lavalink.WithTrack(nextTrack)); err != nil {
			return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: tools.ErrorMsg("Error while playing next track"),
				},
			})
		}
	} else {
		if err := player.Update(context.TODO(), lavalink.WithNullTrack()); err != nil {
			return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: tools.ErrorMsg("Error while stopping the track"),
				},
			})
		}
	}

	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: tools.SuccessMsg("Skipped"),
		},
	})
}

func createPlaylist(event *discordgo.InteractionCreate, bot *Bot) error {
	name := event.ApplicationCommandData().Options[0].StringValue()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := bot.Db.AddUsersPL(ctx, name, event.GuildID); err == db_service.ErrObjectAlreadyExists {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg(fmt.Sprintf("Playlist `%s` already exists", name)),
			},
		})
	} else if err != nil {
		return err
	}

	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: tools.SuccessMsg(fmt.Sprintf("Created playlist `%s`", name)),
		},
	})
}

func deletePlaylist(event *discordgo.InteractionCreate, bot *Bot) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	playlists, err := bot.Db.GetUsersPLs(ctx, event.GuildID)
	if err != nil {
		return fmt.Errorf("error while seraching for users playlists: %v", err)
	}

	if len(playlists) == 0 {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("There are no playlists to delete"),
			},
		})
	}

	options := make([]discordgo.SelectMenuOption, len(playlists))
	for i, p := range playlists {
		options[i] = discordgo.SelectMenuOption{
			Label: p.Name,
			Value: strconv.FormatInt(p.ID, 10),
		}
	}

	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Delete playlist",
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "delete-playlist",
							Placeholder: "Choose playlist",
							Options:     options,
						},
					},
				},
			},
		},
	})
}

func deletePlaylistAction(event *discordgo.InteractionCreate, bot *Bot) error {
	data := event.MessageComponentData()
	id, err := strconv.ParseInt(data.Values[0], 10, 64)
	if err != nil {
		return fmt.Errorf("can't parse playlist id: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = bot.Db.DeletePLByID(ctx, id); err == db_service.ErrObjectNotFound {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg(fmt.Sprintf("There is no such playlist")),
			},
		})
	} else if err != nil {
		return fmt.Errorf("error while deleting playlist: %v", err)
	}

	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: tools.SuccessMsg("Deleted playlist"),
		},
	})
}

func loadPlaylist(event *discordgo.InteractionCreate, bot *Bot) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	playlists, err := bot.Db.GetUsersPLs(ctx, event.GuildID)
	if err != nil {
		return fmt.Errorf("error while seraching for users playlists: %v", err)
	}

	if len(playlists) == 0 {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("There are no playlists to load"),
			},
		})
	}

	options := make([]discordgo.SelectMenuOption, len(playlists))
	for i, p := range playlists {
		options[i] = discordgo.SelectMenuOption{
			Label: p.Name,
			Value: strconv.FormatInt(p.ID, 10),
		}
	}

	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Load playlist",
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "load-playlist",
							Placeholder: "Choose playlist",
							Options:     options,
						},
					},
				},
			},
		},
	})
}

func loadPlaylistAction(event *discordgo.InteractionCreate, bot *Bot) error {
	bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	voiceState, err := bot.Session.State.VoiceState(event.GuildID, event.Member.User.ID)
	if err != nil {
		bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
			Embeds: tools.ErrorMsg("You are not in a voice channel. Please join one and try again"),
		})
		return err
	}

	data := event.MessageComponentData()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(data.Values[0], 10, 64)
	if err != nil {
		return fmt.Errorf("can't parse playlist id: %s", err)
	}

	tracks, err := bot.Db.GetTracksByPlaylist(ctx, id)
	if err == db_service.ErrObjectNotFound {
		bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
			Embeds: tools.ErrorMsg(fmt.Sprintf("There is no such playlist")),
		})
		return err
	} else if err != nil {
		return fmt.Errorf("error while searching for playlist: %v", err)
	}

	tracksToPlay := make([]lavalink.Track, len(tracks))
	for i, track := range tracks {
		tracksToPlay[i] = *track.Value
	}

	queue := bot.Queues.Get(event.GuildID)

	var player disgolink.Player
	player = bot.Lavalink.ExistingPlayer(snowflake.MustParse(event.GuildID))
	if player != nil {
		curTrack := player.Track()
		if curTrack != nil && player.Paused() {
			player.Update(context.TODO(), lavalink.WithPaused(false))
		}
	} else {
		player = bot.Lavalink.Player(snowflake.MustParse(event.GuildID))
	}

	var toPlay *lavalink.Track

	if player.Track() == nil {
		toPlay = &tracksToPlay[0]
		queue.Add(tracksToPlay[1:]...)
	} else {
		queue.Add(tracksToPlay...)
	}

	if toPlay != nil {
		if err = bot.Session.ChannelVoiceJoinManual(event.GuildID, voiceState.ChannelID, false, true); err != nil {
			bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
				Embeds: tools.ErrorMsg("Couldn't join the voice channel"),
			})
			return err
		}

		if err = player.Update(context.TODO(), lavalink.WithTrack(*toPlay)); err != nil {
			return err
		}
	}

	bot.Session.FollowupMessageCreate(event.Interaction, true, &discordgo.WebhookParams{
		Embeds: tools.SuccessMsg("Loaded playlist"),
	})

	return nil
}

func addToPlaylist(event *discordgo.InteractionCreate, bot *Bot) error {
	player := bot.Lavalink.ExistingPlayer(snowflake.MustParse(event.GuildID))
	if player == nil || player.Track() == nil {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("Nothing is being played at the moment"),
			},
		})
	}

	queue := bot.Queues.Get(event.GuildID)
	queue.SaveTrackToAdd(player.Track())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	playlists, err := bot.Db.GetUsersPLs(ctx, event.GuildID)
	if err != nil {
		return fmt.Errorf("error while seraching for users playlists: %v", err)
	}

	if len(playlists) == 0 {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("There are no playlists to add to"),
			},
		})
	}

	options := make([]discordgo.SelectMenuOption, len(playlists))
	for i, p := range playlists {
		options[i] = discordgo.SelectMenuOption{
			Label: p.Name,
			Value: strconv.FormatInt(p.ID, 10),
		}
	}

	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Add to the playlist",
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "add-to-playlist",
							Placeholder: "Choose playlist",
							Options:     options,
						},
					},
				},
			},
		},
	})
}

func addToPlaylistAction(event *discordgo.InteractionCreate, bot *Bot) error {
	queue := bot.Queues.Get(event.GuildID)
	track := queue.GetSavedTrackToAdd()
	if track == nil {
		return fmt.Errorf("there is no saved track to add")
	}

	data := event.MessageComponentData()
	id, err := strconv.ParseInt(data.Values[0], 10, 64)
	if err != nil {
		return fmt.Errorf("can't parse playlist id: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = bot.Db.AddTrack(ctx, &db_service.Track{
		PlaylistID: id,
		Value:      track,
	}); err == db_service.ErrObjectNotFound {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg(fmt.Sprintf("There is no such playlist")),
			},
		})
	} else if err != nil {
		return fmt.Errorf("error while adding track to playlist: %v", err)
	}

	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: tools.SuccessMsg("Added to playlist"),
		},
	})
}

func showPlaylist(event *discordgo.InteractionCreate, bot *Bot) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	playlists, err := bot.Db.GetUsersPLs(ctx, event.GuildID)
	if err != nil {
		return fmt.Errorf("error while seraching for users playlists: %v", err)
	}

	if len(playlists) == 0 {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("There are no playlists to show"),
			},
		})
	}

	options := make([]discordgo.SelectMenuOption, len(playlists))
	for i, p := range playlists {
		options[i] = discordgo.SelectMenuOption{
			Label: p.Name,
			Value: strconv.FormatInt(p.ID, 10),
		}
	}

	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Show playlist",
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "show-playlist",
							Placeholder: "Choose playlist",
							Options:     options,
						},
					},
				},
			},
		},
	})
}

func showPlaylistAction(event *discordgo.InteractionCreate, bot *Bot) error {
	data := event.MessageComponentData()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(data.Values[0], 10, 64)
	if err != nil {
		return fmt.Errorf("can't parse playlist id: %s", err)
	}

	tracks, err := bot.Db.GetTracksByPlaylist(ctx, id)
	if err == db_service.ErrObjectNotFound {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.ErrorMsg("There are no playlists to delete"),
			},
		})
	} else if err != nil {
		return fmt.Errorf("error while searching for playlist: %v", err)
	}

	if len(tracks) == 0 {
		return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: tools.SuccessMsg("No tracks in playlist"),
			},
		})
	}

	var tracksStr string
	for i, track := range tracks {
		tracksStr += fmt.Sprintf("%d. [`%s`](<%s>)\n", i+1, track.Value.Info.Title, *track.Value.Info.URI)
	}

	return bot.Session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: tools.SuccessMsg(fmt.Sprintf("Playlist:\n%s", tracksStr)),
		},
	})
}
