package main

import (
	cfg "erupe-ce/config"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"erupe-ce/common/gametime"
	"erupe-ce/server/api"
	"erupe-ce/server/channelserver"
	"erupe-ce/server/discordbot"
	"erupe-ce/server/entranceserver"
	"erupe-ce/server/signserver"
	"strings"

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

func setupDiscordBot(config *cfg.Config, logger *zap.Logger) *discordbot.DiscordBot {
	bot, err := discordbot.NewDiscordBot(discordbot.Options{
		Logger: logger,
		Config: config,
	})

	if err != nil {
		preventClose(config, fmt.Sprintf("Discord: Failed to start, %s", err.Error()))
	}

	// Discord bot
	err = bot.Start()

	if err != nil {
		preventClose(config, fmt.Sprintf("Discord: Failed to start, %s", err.Error()))
	}

	_, err = bot.Session.ApplicationCommandBulkOverwrite(bot.Session.State.User.ID, "", discordbot.Commands)
	if err != nil {
		preventClose(config, fmt.Sprintf("Discord: Failed to start, %s", err.Error()))
	}

	return bot
}

func main() {
	var err error

	var zapLogger *zap.Logger
	zapLogger, _ = zap.NewDevelopment()

	defer func() { _ = zapLogger.Sync() }()
	logger := zapLogger.Named("main")

	config, cfgErr := cfg.LoadConfig()
	if cfgErr != nil {
		fmt.Println("\nFailed to start Erupe:\n" + fmt.Sprintf("Failed to load config: %s", cfgErr.Error()))
		go wait()
		fmt.Println("\nPress Enter/Return to exit...")
		_, _ = fmt.Scanln()
		os.Exit(0)
	}

	logger.Info(fmt.Sprintf("Starting Erupe (9.3b-%s)", Commit()))
	logger.Info(fmt.Sprintf("Client Mode: %s (%d)", config.ClientMode, config.RealClientMode))

	if config.Database.Password == "" {
		preventClose(config, "Database password is blank")
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
			preventClose(config, "Invalid host address")
		}
	}

	// Discord bot
	var discordBot *discordbot.DiscordBot = nil

	if config.Discord.Enabled {
		discordBot = setupDiscordBot(config, logger)

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
		preventClose(config, fmt.Sprintf("Database: Failed to open, %s", err.Error()))
	}

	// Test the DB connection.
	err = db.Ping()
	if err != nil {
		preventClose(config, fmt.Sprintf("Database: Failed to ping, %s", err.Error()))
	}

	// Configure connection pool to avoid exhausting PostgreSQL under load.
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	logger.Info("Database: Started successfully")

	// Pre-compute all server IDs this instance will own, so we only
	// delete our own rows (safe for multi-instance on the same DB).
	var ownedServerIDs []string
	{
		si := 0
		for _, ee := range config.Entrance.Entries {
			ci := 0
			for range ee.Channels {
				sid := (4096 + si*256) + (16 + ci)
				ownedServerIDs = append(ownedServerIDs, fmt.Sprint(sid))
				ci++
			}
			si++
		}
	}

	// Clear stale data scoped to this instance's server IDs
	if len(ownedServerIDs) > 0 {
		idList := strings.Join(ownedServerIDs, ",")
		if config.DebugOptions.ProxyPort == 0 {
			_ = db.MustExec("DELETE FROM sign_sessions WHERE server_id IN (" + idList + ")")
		}
		_ = db.MustExec("DELETE FROM servers WHERE server_id IN (" + idList + ")")
	}
	_ = db.MustExec(`UPDATE guild_characters SET treasure_hunt=NULL`)

	// Clean the DB if the option is on.
	if config.DebugOptions.CleanDB {
		logger.Info("Database: Started clearing...")
		cleanDB(db)
		logger.Info("Database: Finished clearing")
	}

	logger.Info(fmt.Sprintf("Server Time: %s", gametime.Adjusted().String()))

	// Now start our server(s).

	// Entrance server.

	var entranceServer *entranceserver.Server
	if config.Entrance.Enabled {
		entranceServer = entranceserver.NewServer(
			&entranceserver.Config{
				Logger:      logger.Named("entrance"),
				ErupeConfig: config,
				DB:          db,
			})
		err = entranceServer.Start()
		if err != nil {
			preventClose(config, fmt.Sprintf("Entrance: Failed to start, %s", err.Error()))
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
				ErupeConfig: config,
				DB:          db,
			})
		err = signServer.Start()
		if err != nil {
			preventClose(config, fmt.Sprintf("Sign: Failed to start, %s", err.Error()))
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
				ErupeConfig: config,
				DB:          db,
			})
		err = ApiServer.Start()
		if err != nil {
			preventClose(config, fmt.Sprintf("API: Failed to start, %s", err.Error()))
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
				if !ce.IsEnabled() {
					logger.Info(fmt.Sprintf("Channel %d (%d): Disabled via config", count, ce.Port))
					ci++
					count++
					continue
				}
				c := *channelserver.NewServer(&channelserver.Config{
					ID:          uint16(sid),
					Logger:      logger.Named("channel-" + fmt.Sprint(count)),
					ErupeConfig: config,
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
					preventClose(config, fmt.Sprintf("Channel: Failed to start, %s", err.Error()))
				} else {
					channelQuery += fmt.Sprintf(
						`INSERT INTO servers (server_id, current_players, world_name, world_description, land) VALUES (%d, 0, '%s', '%s', %d);`,
						sid, ee.Name, ee.Description, i+1,
					)
					channels = append(channels, &c)
					logger.Info(fmt.Sprintf("Channel %d (%d): Started successfully", count, ce.Port))
					count++
				}
				ci++
			}
			ci = 0
			si++
		}

		// Register all servers in DB
		_ = db.MustExec(channelQuery)

		registry := channelserver.NewLocalChannelRegistry(channels)
		for _, c := range channels {
			c.Channels = channels
			c.Registry = registry
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

func preventClose(config *cfg.Config, text string) {
	if config != nil && config.DisableSoftCrash {
		os.Exit(0)
	}
	fmt.Println("\nFailed to start Erupe:\n" + text)
	go wait()
	fmt.Println("\nPress Enter/Return to exit...")
	_, _ = fmt.Scanln()
	os.Exit(0)
}
