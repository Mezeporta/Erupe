package main

import (
	_config "erupe-ce/config"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"erupe-ce/server/api"
	"erupe-ce/server/channelserver"
	"erupe-ce/server/discordbot"
	"erupe-ce/server/entranceserver"
	"erupe-ce/server/signserver"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Temporary DB auto clean on startup for quick development & testing.
func cleanDB(db *sqlx.DB) {
	_ = db.MustExec("DELETE FROM guild_characters")
	_ = db.MustExec("DELETE FROM guilds")
	_ = db.MustExec("DELETE FROM characters")
	_ = db.MustExec("DELETE FROM users")
}

var Commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value[:7]
			}
		}
	}
	return "unknown"
}

func setupDiscordBot(db *discordbot.DiscordBot, logger *zap.Logger) {
	bot, err := discordbot.NewDiscordBot(discordbot.Options{
		Logger: logger,
		Config: _config.ErupeConfig,
	})

	if err != nil {
		preventClose(fmt.Sprintf("Discord: Failed to start, %s", err.Error()))
	}

	// Discord bot
	err = bot.Start()

	if err != nil {
		preventClose(fmt.Sprintf("Discord: Failed to start, %s", err.Error()))
	}

	*db = *bot

	_, err = bot.Session.ApplicationCommandBulkOverwrite(bot.Session.State.User.ID, "", discordbot.Commands)
	if err != nil {
		preventClose(fmt.Sprintf("Discord: Failed to start, %s", err.Error()))
	}
}

func main() {
	var err error

	var zapLogger *zap.Logger
	config := _config.ErupeConfig
	zapLogger, _ = zap.NewDevelopment()

	defer zapLogger.Sync()
	logger := zapLogger.Named("main")

	logger.Info(fmt.Sprintf("Starting Erupe (9.3b-%s)", Commit()))
	logger.Info(fmt.Sprintf("Client Mode: %s (%d)", config.ClientMode, config.RealClientMode))

	if config.Database.Password == "" {
		preventClose("Database password is blank")
	}

	if net.ParseIP(config.Host) == nil {
		ips, _ := net.LookupIP(config.Host)
		for _, ip := range ips {
			if ip != nil {
				config.Host = ip.String()
				break
			}
		}
		if net.ParseIP(config.Host) == nil {
			preventClose("Invalid host address")
		}
	}

	// Discord bot
	var discordBot *discordbot.DiscordBot = nil

	if config.Discord.Enabled {
		setupDiscordBot(discordBot, logger)

		logger.Info("Discord: Started successfully")
	} else {
		logger.Info("Discord: Disabled")
	}

	// Create the postgres DB pool.
	connectString := fmt.Sprintf(
		"host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode=disable",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Database,
	)

	db, err := sqlx.Open("postgres", connectString)
	if err != nil {
		preventClose(fmt.Sprintf("Database: Failed to open, %s", err.Error()))
	}

	// Test the DB connection.
	err = db.Ping()
	if err != nil {
		preventClose(fmt.Sprintf("Database: Failed to ping, %s", err.Error()))
	}
	logger.Info("Database: Started successfully")

	// Clear stale data
	if config.DebugOptions.ProxyPort == 0 {
		_ = db.MustExec("DELETE FROM sign_sessions")
	}
	_ = db.MustExec("DELETE FROM servers")
	_ = db.MustExec(`UPDATE guild_characters SET treasure_hunt=NULL`)

	// Clean the DB if the option is on.
	if config.DebugOptions.CleanDB {
		logger.Info("Database: Started clearing...")
		cleanDB(db)
		logger.Info("Database: Finished clearing")
	}

	logger.Info(fmt.Sprintf("Server Time: %s", channelserver.TimeAdjusted().String()))

	// Now start our server(s).

	// Entrance server.

	var entranceServer *entranceserver.Server
	if config.Entrance.Enabled {
		entranceServer = entranceserver.NewServer(
			&entranceserver.Config{
				Logger:      logger.Named("entrance"),
				ErupeConfig: _config.ErupeConfig,
				DB:          db,
			})
		err = entranceServer.Start()
		if err != nil {
			preventClose(fmt.Sprintf("Entrance: Failed to start, %s", err.Error()))
		}
		logger.Info("Entrance: Started successfully")
	} else {
		logger.Info("Entrance: Disabled")
	}

	// Sign server.

	var signServer *signserver.Server
	if config.Sign.Enabled {
		signServer = signserver.NewServer(
			&signserver.Config{
				Logger:      logger.Named("sign"),
				ErupeConfig: _config.ErupeConfig,
				DB:          db,
			})
		err = signServer.Start()
		if err != nil {
			preventClose(fmt.Sprintf("Sign: Failed to start, %s", err.Error()))
		}
		logger.Info("Sign: Started successfully")
	} else {
		logger.Info("Sign: Disabled")
	}

	// New Sign server
	var ApiServer *api.APIServer
	if config.API.Enabled {
		ApiServer = api.NewAPIServer(
			&api.Config{
				Logger:      logger.Named("sign"),
				ErupeConfig: _config.ErupeConfig,
				DB:          db,
			})
		err = ApiServer.Start()
		if err != nil {
			preventClose(fmt.Sprintf("API: Failed to start, %s", err.Error()))
		}
		logger.Info("API: Started successfully")
	} else {
		logger.Info("API: Disabled")
	}

	var channels []*channelserver.Server

	if config.Channel.Enabled {
		channelQuery := ""
		si := 0
		ci := 0
		count := 1
		for j, ee := range config.Entrance.Entries {
			for i, ce := range ee.Channels {
				sid := (4096 + si*256) + (16 + ci)
				c := *channelserver.NewServer(&channelserver.Config{
					ID:          uint16(sid),
					Logger:      logger.Named("channel-" + fmt.Sprint(count)),
					ErupeConfig: _config.ErupeConfig,
					DB:          db,
					DiscordBot:  discordBot,
				})
				if ee.IP == "" {
					c.IP = config.Host
				} else {
					c.IP = ee.IP
				}
				c.Port = ce.Port
				c.GlobalID = fmt.Sprintf("%02d%02d", j+1, i+1)
				err = c.Start()
				if err != nil {
					preventClose(fmt.Sprintf("Channel: Failed to start, %s", err.Error()))
				} else {
					channelQuery += fmt.Sprintf(
						`INSERT INTO servers (server_id, current_players, world_name, world_description, land) VALUES (%d, 0, '%s', '%s', %d);`,
						sid, ee.Name, ee.Description, i+1
					)
					channels = append(channels, &c)
					logger.Info(fmt.Sprintf("Channel %d (%d): Started successfully", count, ce.Port))
					ci++
					count++
				}
			}
			ci = 0
			si++
		}

		// Register all servers in DB
		_ = db.MustExec(channelQuery)

		for _, c := range channels {
			c.Channels = channels
		}
	}

	logger.Info("Finished starting Erupe")

	// Wait for exit or interrupt with ctrl+C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	if !config.DisableSoftCrash {
		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("Shutting down in %d...", 10-i)
			for _, c := range channels {
				c.BroadcastChatMessage(message)
			}
			logger.Info(message)
			time.Sleep(time.Second)
		}
	}

	if config.Channel.Enabled {
		for _, c := range channels {
			c.Shutdown()
		}
	}

	if config.Sign.Enabled {
		signServer.Shutdown()
	}

	if config.API.Enabled {
		ApiServer.Shutdown()
	}

	if config.Entrance.Enabled {
		entranceServer.Shutdown()
	}

	time.Sleep(1 * time.Second)
}

func wait() {
	for {
		time.Sleep(time.Millisecond * 100)
	}
}

func preventClose(text string) {
	if _config.ErupeConfig.DisableSoftCrash {
		os.Exit(0)
	}
	fmt.Println("\nFailed to start Erupe:\n" + text)
	go wait()
	fmt.Println("\nPress Enter/Return to exit...")
	fmt.Scanln()
	os.Exit(0)
}
