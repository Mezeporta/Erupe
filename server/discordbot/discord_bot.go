package discordbot

import (
	_config "erupe-ce/config"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// Commands defines the slash commands registered with Discord, including
// account linking and password management.
var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "link",
		Description: "Link your Erupe account to Discord",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "token",
				Description: "The token provided by the Discord command in-game",
				Required:    true,
			},
		},
	},
	{
		Name:        "password",
		Description: "Change your Erupe account password",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "password",
				Description: "Your new password",
				Required:    true,
			},
		},
	},
}

// DiscordBot manages a Discord session and provides methods for relaying
// messages between the game server and a configured Discord channel.
type DiscordBot struct {
	Session      *discordgo.Session
	config       *_config.Config
	logger       *zap.Logger
	MainGuild    *discordgo.Guild
	RelayChannel *discordgo.Channel
}

// Options holds the configuration and logger required to create a DiscordBot.
type Options struct {
	Config *_config.Config
	Logger *zap.Logger
}

// NewDiscordBot creates a DiscordBot using the provided options, establishing
// a Discord session and optionally resolving the relay channel.
func NewDiscordBot(options Options) (discordBot *DiscordBot, err error) {
	session, err := discordgo.New("Bot " + options.Config.Discord.BotToken)

	if err != nil {
		options.Logger.Fatal("Discord failed", zap.Error(err))
		return nil, err
	}

	var relayChannel *discordgo.Channel

	if options.Config.Discord.RelayChannel.Enabled {
		relayChannel, err = session.Channel(options.Config.Discord.RelayChannel.RelayChannelID)
	}

	if err != nil {
		options.Logger.Fatal("Discord failed to create relayChannel", zap.Error(err))
		return nil, err
	}

	discordBot = &DiscordBot{
		config:       options.Config,
		logger:       options.Logger,
		Session:      session,
		RelayChannel: relayChannel,
	}

	return
}

// Start opens the websocket connection to Discord.
func (bot *DiscordBot) Start() (err error) {
	err = bot.Session.Open()

	return
}

// NormalizeDiscordMessage replaces all mentions to real name from the message.
func (bot *DiscordBot) NormalizeDiscordMessage(message string) string {
	userRegex := regexp.MustCompile(`<@!?(\d{17,19})>`)
	emojiRegex := regexp.MustCompile(`(?:<a?)?:(\w+):(?:\d{18}>)?`)

	result := ReplaceTextAll(message, userRegex, func(userId string) string {
		user, err := bot.Session.User(userId)

		if err != nil {
			return "@unknown" // @Unknown
		}

		return "@" + user.Username
	})

	result = ReplaceTextAll(result, emojiRegex, func(emojiName string) string {
		return ":" + emojiName + ":"
	})

	return result
}

// RealtimeChannelSend sends a message to the configured relay channel. If no
// relay channel is configured, the call is a no-op.
func (bot *DiscordBot) RealtimeChannelSend(message string) (err error) {
	if bot.RelayChannel == nil {
		return
	}

	_, err = bot.Session.ChannelMessageSend(bot.RelayChannel.ID, message)

	return
}
// ReplaceTextAll replaces every match of regex in text by calling handler with
// the first capture group of each match and substituting the result.
func ReplaceTextAll(text string, regex *regexp.Regexp, handler func(input string) string) string {
	result := regex.ReplaceAllFunc([]byte(text), func(s []byte) []byte {
		input := regex.ReplaceAllString(string(s), `$1`)

		return []byte(handler(input))
	})

	return string(result)
}
