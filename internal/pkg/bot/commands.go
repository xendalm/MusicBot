package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type CommandHandler func(event *discordgo.InteractionCreate, bot *Bot) error

type Command struct {
	Command *discordgo.ApplicationCommand
	Handler CommandHandler
}

func NewCommand(command *discordgo.ApplicationCommand, handler CommandHandler) *Command {
	return &Command{
		Command: command,
		Handler: handler,
	}
}

func CreateCommands(s *discordgo.Session, guildID string) {
	for _, c := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, c.Command)
		if err != nil {
			fmt.Println("Error while creating command: ", err)
		}
		fmt.Println("Created command: ", cmd.Name, cmd.ID)
	}
	for _, c := range components {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, c.Command)
		if err != nil {
			fmt.Println("Error while creating component command: ", err)
		}
		fmt.Println("Created component command: ", cmd.Name, cmd.ID)
	}
}

func DeleteCommands(s *discordgo.Session, guildID string) {
	appCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		fmt.Println("Error while finding commands to delete: ", err)
		return
	}
	for _, c := range appCommands {
		if err = s.ApplicationCommandDelete(s.State.User.ID, guildID, c.ID); err != nil {
			fmt.Println("Error while deleting command: ", err)
			return
		}
		fmt.Println("Deleted command: ", c.Name)
	}
}

func InteractionReceived(s *discordgo.Session, i *discordgo.InteractionCreate, bot *Bot) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if c, ok := commands[i.ApplicationCommandData().Name]; ok {
			err := c.Handler(i, bot)
			if err != nil {
				fmt.Printf("Error while handling command interaction [%v]\n", err)
			}
		}
	case discordgo.InteractionMessageComponent:
		if c, ok := components[i.MessageComponentData().CustomID]; ok {
			err := c.Handler(i, bot)
			if err != nil {
				fmt.Printf("Error while handling component interaction [%v]\n", err)
			}
		}
	}
}

var commands = map[string]*Command{
	"ping": NewCommand(&discordgo.ApplicationCommand{
		Name:        "ping",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Pong!",
	}, ping),
	"play": NewCommand(&discordgo.ApplicationCommand{
		Name:        "play",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Plays a song",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "identifier",
				Description: "The song link or search query",
				Required:    true,
			},
		},
	}, play),
	"pause": NewCommand(&discordgo.ApplicationCommand{
		Name:        "pause",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Pauses/resumes the current song",
	}, pause),
	"stop": NewCommand(&discordgo.ApplicationCommand{
		Name:        "stop",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Stops the current song and player",
	}, stop),
	"nowplaying": NewCommand(&discordgo.ApplicationCommand{
		Name:        "nowplaying",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Shows the current playing song",
	}, nowPlaying),
	"queue": NewCommand(&discordgo.ApplicationCommand{
		Name:        "queue",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Shows the current queue",
	}, showQueue),
	"clear": NewCommand(&discordgo.ApplicationCommand{
		Name:        "clear",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Shows the current queue",
	}, clearQueue),
	"shuffle": NewCommand(&discordgo.ApplicationCommand{
		Name:        "shuffle",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Shuffles the current queue",
	}, shuffle),
	"skip": NewCommand(&discordgo.ApplicationCommand{
		Name:        "skip",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Skips the current song",
	}, skip),
	"create-playlist": NewCommand(&discordgo.ApplicationCommand{
		Name:        "create-playlist",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Creates new playlist",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "name",
				Description: "Name",
				Required:    true,
			},
		},
	}, createPlaylist),
	"delete-playlist": NewCommand(&discordgo.ApplicationCommand{
		Name:        "delete-playlist",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Deletes the playlist",
	}, deletePlaylist),
	"load-playlist": NewCommand(&discordgo.ApplicationCommand{
		Name:        "load-playlist",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Adds the playlist to the queue",
	}, loadPlaylist),
	"add-to-playlist": NewCommand(&discordgo.ApplicationCommand{
		Name:        "add-to-playlist",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Adds the playing track to the playlist",
	}, addToPlaylist),
	"show-playlist": NewCommand(&discordgo.ApplicationCommand{
		Name:        "show-playlist",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Shows the playlists content",
	}, showPlaylist),
}

var components = map[string]*Command{
	"delete-playlist": NewCommand(&discordgo.ApplicationCommand{
		Name:        "delete-playlist",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Deletes the playlist",
	}, deletePlaylistAction),
	"load-playlist": NewCommand(&discordgo.ApplicationCommand{
		Name:        "load-playlist",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Adds the playlist to the queue",
	}, loadPlaylistAction),
	"add-to-playlist": NewCommand(&discordgo.ApplicationCommand{
		Name:        "add-to-playlist",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Adds the playing track to the playlist",
	}, addToPlaylistAction), "show-playlist": NewCommand(&discordgo.ApplicationCommand{
		Name:        "show-playlist",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Shows the playlists content",
	}, showPlaylistAction),
}
