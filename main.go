package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"

	"erupe-ce/config"
	"erupe-ce/server/channelserver"
	"erupe-ce/server/discordbot"
	"erupe-ce/server/entranceserver"
	"erupe-ce/server/signserver"
	"erupe-ce/server/signv2server"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Temporary DB auto clean on startup for quick development & testing.
func cleanDB(db *sqlx.DB) {
	_ = db.MustExec("DELETE FROM guild_characters")
	_ = db.MustExec("DELETE FROM guilds")
	_ = db.MustExec("DELETE FROM characters")
	_ = db.MustExec("DELETE FROM sign_sessions")
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

// initLogger initializes the zap logger with file logging support
func initLogger(cfg *config.Config) *zap.Logger {
	var zapConfig zap.Config

	// Base configuration based on DevMode
	if cfg.DevMode {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Configure output paths
	if cfg.Logging.LogToFile {
		// Ensure the log directory exists
		logDir := filepath.Dir(cfg.Logging.LogFilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("Failed to create log directory: %s\n", err)
			os.Exit(1)
		}

		// Create lumberjack logger for file rotation
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.Logging.LogFilePath,
			MaxSize:    cfg.Logging.LogMaxSize,
			MaxBackups: cfg.Logging.LogMaxBackups,
			MaxAge:     cfg.Logging.LogMaxAge,
			Compress:   cfg.Logging.LogCompress,
		}

		// Create encoder
		encoderConfig := zapConfig.EncoderConfig
		var encoder zapcore.Encoder
		if cfg.DevMode {
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		} else {
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		}

		// Create cores for both console and file output
		consoleCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stderr),
			zapConfig.Level,
		)
		fileCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(fileWriter),
			zapConfig.Level,
		)

		// Combine cores
		core := zapcore.NewTee(consoleCore, fileCore)
		logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		return logger
	}

	// File logging disabled, use default configuration
	logger, err := zapConfig.Build()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %s\n", err)
		os.Exit(1)
	}
	return logger
}

func main() {
	var err error

	// Initialize logger with file support
	zapLogger := initLogger(config.ErupeConfig)
	defer zapLogger.Sync()
	logger := zapLogger.Named("main")

	logger.Info(fmt.Sprintf("Starting Erupe (9.2-%s)", Commit()))

	if config.ErupeConfig.Database.Password == "" {
		preventClose("Database password is blank")
	}

	if net.ParseIP(config.ErupeConfig.Host) == nil {
		ips, _ := net.LookupIP(config.ErupeConfig.Host)
		for _, ip := range ips {
			if ip != nil {
				config.ErupeConfig.Host = ip.String()
				break
			}
		}
		if net.ParseIP(config.ErupeConfig.Host) == nil {
			preventClose("Invalid host address")
		}
	}

	// Discord bot
	var discordBot *discordbot.DiscordBot = nil

	if config.ErupeConfig.Discord.Enabled {
		bot, err := discordbot.NewDiscordBot(discordbot.Options{
			Logger: logger,
			Config: config.ErupeConfig,
		})

		if err != nil {
			preventClose(fmt.Sprintf("Discord: Failed to start, %s", err.Error()))
		}

		// Discord bot
		err = bot.Start()

		if err != nil {
			preventClose(fmt.Sprintf("Discord: Failed to start, %s", err.Error()))
		}

		discordBot = bot
		logger.Info("Discord: Started successfully")
	} else {
		logger.Info("Discord: Disabled")
	}

	// Create the postgres DB pool.
	connectString := fmt.Sprintf(
		"host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode=disable",
		config.ErupeConfig.Database.Host,
		config.ErupeConfig.Database.Port,
		config.ErupeConfig.Database.User,
		config.ErupeConfig.Database.Password,
		config.ErupeConfig.Database.Database,
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
	_ = db.MustExec("DELETE FROM sign_sessions")
	_ = db.MustExec("DELETE FROM servers")

	// Clean the DB if the option is on.
	if config.ErupeConfig.DevMode && config.ErupeConfig.DevModeOptions.CleanDB {
		logger.Info("Database: Started clearing...")
		cleanDB(db)
		logger.Info("Database: Finished clearing")
	}

	logger.Info(fmt.Sprintf("Server Time: %s", channelserver.TimeAdjusted().String()))

	// Now start our server(s).

	// Entrance server.

	var entranceServer *entranceserver.Server
	if config.ErupeConfig.Entrance.Enabled {
		entranceServer = entranceserver.NewServer(
			&entranceserver.Config{
				Logger:      logger.Named("entrance"),
				ErupeConfig: config.ErupeConfig,
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
	if config.ErupeConfig.Sign.Enabled {
		signServer = signserver.NewServer(
			&signserver.Config{
				Logger:      logger.Named("sign"),
				ErupeConfig: config.ErupeConfig,
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
	var newSignServer *signv2server.Server
	if config.ErupeConfig.SignV2.Enabled {
		newSignServer = signv2server.NewServer(
			&signv2server.Config{
				Logger:      logger.Named("sign"),
				ErupeConfig: config.ErupeConfig,
				DB:          db,
			})
		err = newSignServer.Start()
		if err != nil {
			preventClose(fmt.Sprintf("SignV2: Failed to start, %s", err.Error()))
		}
		logger.Info("SignV2: Started successfully")
	} else {
		logger.Info("SignV2: Disabled")
	}

	var channels []*channelserver.Server

	if config.ErupeConfig.Channel.Enabled {
		channelQuery := ""
		si := 0
		ci := 0
		count := 1
		for j, ee := range config.ErupeConfig.Entrance.Entries {
			for i, ce := range ee.Channels {
				sid := (4096 + si*256) + (16 + ci)
				c := *channelserver.NewServer(&channelserver.Config{
					ID:          uint16(sid),
					Logger:      logger.Named("channel-" + fmt.Sprint(count)),
					ErupeConfig: config.ErupeConfig,
					DB:          db,
					DiscordBot:  discordBot,
				})
				if ee.IP == "" {
					c.IP = config.ErupeConfig.Host
				} else {
					c.IP = ee.IP
				}
				c.Port = ce.Port
				c.GlobalID = fmt.Sprintf("%02d%02d", j+1, i+1)
				err = c.Start()
				if err != nil {
					preventClose(fmt.Sprintf("Channel: Failed to start, %s", err.Error()))
				} else {
					channelQuery += fmt.Sprintf(`INSERT INTO servers (server_id, season, current_players, world_name, world_description, land) VALUES (%d, %d, 0, '%s', '%s', %d);`, sid, si%3, ee.Name, ee.Description, i+1)
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

	if !config.ErupeConfig.DisableSoftCrash {
		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("Shutting down in %d...", 10-i)
			for _, c := range channels {
				c.BroadcastChatMessage(message)
			}
			logger.Info(message)
			time.Sleep(time.Second)
		}
	}

	if config.ErupeConfig.Channel.Enabled {
		for _, c := range channels {
			c.Shutdown()
		}
	}

	if config.ErupeConfig.Sign.Enabled {
		signServer.Shutdown()
	}

	if config.ErupeConfig.SignV2.Enabled {
		newSignServer.Shutdown()
	}

	if config.ErupeConfig.Entrance.Enabled {
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
	if config.ErupeConfig.DisableSoftCrash {
		os.Exit(0)
	}
	fmt.Println("\nFailed to start Erupe:\n" + text)
	go wait()
	fmt.Println("\nPress Enter/Return to exit...")
	fmt.Scanln()
	os.Exit(0)
}
