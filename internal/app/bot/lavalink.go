package bot

import (
	"bobot/internal/config"
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"time"
)

func (b *Bot) NewLavalinkClient() {
	b.Lavalink = disgolink.New(snowflake.MustParse(b.Session.State.User.ID),
		disgolink.WithListenerFunc(b.OnPlayerPause),
		disgolink.WithListenerFunc(b.OnPlayerResume),
		disgolink.WithListenerFunc(b.OnTrackStart),
		disgolink.WithListenerFunc(b.OnTrackEnd),
		disgolink.WithListenerFunc(b.OnTrackException),
		disgolink.WithListenerFunc(b.OnTrackStuck),
		disgolink.WithListenerFunc(b.OnWebSocketClosed),
	)
}

func (b *Bot) AddLavalinkNode(cfg config.BotConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	node, err := b.Lavalink.AddNode(ctx, disgolink.NodeConfig{
		Name:     cfg.NodeName,
		Address:  cfg.NodeHost + ":" + cfg.NodePort,
		Password: cfg.NodePassword,
		Secure:   cfg.NodeSecure,
	})
	if err != nil {
		log.Fatal(err)
	}

	version, err := node.Version(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("node version: %s", version)
}

func (b *Bot) onVoiceStateUpdate(session *discordgo.Session, event *discordgo.VoiceStateUpdate) {
	if event.UserID != session.State.User.ID {
		return
	}

	var channelID *snowflake.ID
	if event.ChannelID != "" {
		id := snowflake.MustParse(event.ChannelID)
		channelID = &id
	}
	b.Lavalink.OnVoiceStateUpdate(context.TODO(), snowflake.MustParse(event.VoiceState.GuildID), channelID, event.SessionID)
}

func (b *Bot) onVoiceServerUpdate(_ *discordgo.Session, event *discordgo.VoiceServerUpdate) {
	b.Lavalink.OnVoiceServerUpdate(context.TODO(), snowflake.MustParse(event.GuildID), event.Token, event.Endpoint)
}
