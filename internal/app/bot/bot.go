package bot

import (
	"bobot/internal/app/bot/db_service"
	"bobot/internal/app/bot/tools"
	"bobot/internal/app/db"
	"bobot/internal/config"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/caarlos0/env/v6"
	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/log"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

var (
	urlPattern    = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
	searchPattern = regexp.MustCompile(`^(.{2})search:(.+)`)
)

type Bot struct {
	Session  *discordgo.Session
	Lavalink disgolink.Client
	Queues   *QueueManager
	Db       *db_service.DbService
}

func NewBot(db db.DB) *Bot {
	return &Bot{
		Queues: &QueueManager{
			queues: make(map[string]*Queue),
		},
		Db: db_service.NewDbService(db),
	}
}

func (b *Bot) Run() {
	cfg := config.BotConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("unable to parse BotConfig: %v", err)
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetLevel(log.LevelInfo)
	log.Info("Starting...")
	log.Info("discordgo version: ", discordgo.VERSION)
	log.Info("disgolink version: ", disgolink.Version)

	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		log.Fatal(err)
	}

	b.Session = session
	session.State.TrackVoice = true
	session.Identify.Intents = discordgo.IntentGuilds | discordgo.IntentsGuildVoiceStates
	session.AddHandler(b.onVoiceStateUpdate)
	session.AddHandler(b.onVoiceServerUpdate)
	session.AddHandler(b.onGuildCreate)
	session.AddHandler(b.onGuildDelete)
	if err = session.Open(); err != nil {
		log.Fatal(err)
	}
	defer func(session *discordgo.Session) {
		err = session.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(session)

	b.NewLavalinkClient()
	b.AddLavalinkNode(cfg)

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		InteractionReceived(s, i, b)
	})

	log.Info("Bot is now running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s

}

func (b *Bot) onGuildCreate(session *discordgo.Session, event *discordgo.GuildCreate) {
	CreateCommands(session, event.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := b.Db.AddUser(ctx, event.ID); err != nil {
		log.Fatal(err)
	}

	fmt.Println("!!! Guild Created", event.ID)

	for _, channel := range event.Channels {
		if channel.Type == discordgo.ChannelTypeGuildText {
			session.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{Embeds: tools.SuccessMsg("Ready!")})
		}
	}
}

func (b *Bot) onGuildDelete(_ *discordgo.Session, event *discordgo.GuildDelete) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := b.Db.DeleteUser(ctx, event.ID); err != nil {
		log.Fatal(err)
	}

	fmt.Println("!!! Guild Deleted", event.ID)
}
