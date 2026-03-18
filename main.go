package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	cfg "erupe-ce/config"
	"erupe-ce/common/gametime"
	"erupe-ce/server/api"
	"erupe-ce/server/channelserver"
	"erupe-ce/server/discordbot"
	"erupe-ce/server/entranceserver"
	"erupe-ce/server/migrations"
	"erupe-ce/server/setup"
	"erupe-ce/server/signserver"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

	if err = bot.RegisterCommands(); err != nil {
		preventClose(config, fmt.Sprintf("Discord: Failed to start, %s", err.Error()))
	}

	return bot
}

func main() {
	runSetup := flag.Bool("setup", false, "Launch the setup wizard (even if config.json exists)")
	flag.Parse()

	var err error

	var zapLogger *zap.Logger
	zapLogger, _ = zap.NewDevelopment()

	defer func() { _ = zapLogger.Sync() }()
	logger := zapLogger.Named("main")

	if *runSetup {
		logger.Info("Launching setup wizard (--setup)")
		if err := setup.Run(logger.Named("setup"), 8080); err != nil {
			logger.Fatal("Setup wizard failed", zap.Error(err))
		}
	}

	config, cfgErr := cfg.LoadConfig()
	if cfgErr != nil {
		if _, err := os.Stat("config.json"); os.IsNotExist(err) {
			logger.Info("No config.json found, launching setup wizard")
			if err := setup.Run(logger.Named("setup"), 8080); err != nil {
				logger.Fatal("Setup wizard failed", zap.Error(err))
			}
			config, cfgErr = cfg.LoadConfig()
			if cfgErr != nil {
				logger.Fatal("Config still invalid after setup", zap.Error(cfgErr))
			}
		} else {
			preventClose(config, fmt.Sprintf("Failed to load config: %s", cfgErr.Error()))
		}
	}

	logger.Info(fmt.Sprintf("Starting Erupe (9.3.0-%s)", Commit()))
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
		preventClose(config, fmt.Sprintf(
			"Database: Failed to open connection to %s:%d\nError: %s",
			config.Database.Host, config.Database.Port, err.Error(),
		))
	}

	// Test the DB connection.
	err = db.Ping()
	if err != nil {
		dbAddr := fmt.Sprintf("%s:%d", config.Database.Host, config.Database.Port)
		var hint string
		var pqErr *pq.Error
		var netErr *net.OpError
		switch {
		case errors.As(err, &netErr):
			hint = fmt.Sprintf(
				"Database: PostgreSQL is not reachable at %s\n"+
					"  Check that PostgreSQL is running: systemctl status postgresql (or docker ps)\n"+
					"  Using Docker? Run: cd docker && docker compose up db -d",
				dbAddr,
			)
		case errors.As(err, &pqErr) && pqErr.Code == "28P01":
			hint = fmt.Sprintf(
				"Database: Wrong password for user '%s' at %s\n"+
					"  Update Database.Password in config.json to match your PostgreSQL password.",
				config.Database.User, dbAddr,
			)
		case errors.As(err, &pqErr) && pqErr.Code == "3D000":
			hint = fmt.Sprintf(
				"Database: Database '%s' does not exist on %s\n"+
					"  Create it with: createdb -U %s %s",
				config.Database.Database, dbAddr,
				config.Database.User, config.Database.Database,
			)
		default:
			hint = fmt.Sprintf("Database: Failed to connect to %s — %s", dbAddr, err.Error())
		}
		preventClose(config, hint)
	}

	// Configure connection pool to avoid exhausting PostgreSQL under load.
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	logger.Info("Database: Started successfully")

	// Run database migrations
	verBefore, _ := migrations.Version(db)
	applied, migErr := migrations.Migrate(db, logger.Named("migrations"))
	if migErr != nil {
		preventClose(config, fmt.Sprintf("Database migration failed: %s", migErr.Error()))
	}
	if applied > 0 {
		ver, _ := migrations.Version(db)
		logger.Info(fmt.Sprintf("Database: Applied %d migration(s), now at version %d", applied, ver))
	}

	// Auto-apply seed data on a fresh database so users who skip the wizard
	// still get shops, events, and gacha. Seed files use ON CONFLICT DO NOTHING
	// so this is safe to run even if data already exists.
	if verBefore == 0 && applied > 0 {
		seedApplied, seedErr := migrations.ApplySeedData(db, logger.Named("migrations"))
		if seedErr != nil {
			logger.Warn(fmt.Sprintf("Seed data failed: %s", seedErr.Error()))
		} else if seedApplied > 0 {
			logger.Info(fmt.Sprintf("Database: Applied %d seed data file(s)", seedApplied))
		}
	}

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
			if _, err := db.Exec("DELETE FROM sign_sessions WHERE server_id IN (" + idList + ")"); err != nil {
				logger.Warn("Failed to clear stale sign sessions", zap.Error(err))
			}
		}
		if _, err := db.Exec("DELETE FROM servers WHERE server_id IN (" + idList + ")"); err != nil {
			logger.Warn("Failed to clear stale server entries", zap.Error(err))
		}
	}
	if _, err := db.Exec(`UPDATE guild_characters SET treasure_hunt=NULL`); err != nil {
		logger.Warn("Failed to reset treasure hunts", zap.Error(err))
	}

	// Clean the DB if the option is on.
	if config.DebugOptions.CleanDB {
		logger.Info("Database: Started clearing...")
		cleanDB(db)
		logger.Info("Database: Finished clearing")
	}

	logger.Info(fmt.Sprintf("Server Time: %s", gametime.Adjusted().String()))

	// Warn if quest files are missing — clients crash without them.
	questPath := filepath.Join(config.BinPath, "quests")
	if entries, err := os.ReadDir(questPath); err != nil || len(entries) == 0 {
		logger.Warn("No quest files found in " + questPath)
		logger.Warn("Download quest/scenario files from: https://files.catbox.moe/xf0l7w.7z")
		logger.Warn("Extract into your BinPath directory (default: bin/)")
		logger.Warn("Without these files, quests will not load and clients will crash.")
	}

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
			c.Registry = registry
		}
	}

	logger.Info("Finished starting Erupe")

	// Wait for exit or interrupt with ctrl+C.
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	if !config.DisableSoftCrash {
		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("Shutting down in %d...", 10-i)
			for _, c := range channels {
				c.BroadcastChatMessage(message)
			}
			logger.Info(message)
			select {
			case <-sig:
				logger.Info("Second signal received, forcing shutdown")
				os.Exit(1)
			case <-time.After(time.Second):
			}
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
