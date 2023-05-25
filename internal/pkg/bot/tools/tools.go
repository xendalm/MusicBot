package tools

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v2/lavalink"
)

func ErrorMsg(msg string) []*discordgo.MessageEmbed {
	embeds := []*discordgo.MessageEmbed{
		{
			Type:        "rich",
			Color:       0xf44336,
			Description: msg,
		},
	}
	return embeds
}

func SuccessMsg(msg string) []*discordgo.MessageEmbed {
	embeds := []*discordgo.MessageEmbed{
		{
			Type:        "rich",
			Color:       0x61ca33,
			Description: msg,
		},
	}
	return embeds
}

func FormatPosition(position lavalink.Duration) string {
	if position == 0 {
		return "0:00"
	}
	return fmt.Sprintf("%d:%02d", position.Minutes(), position.SecondsPart())
}
