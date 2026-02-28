package main

import (
	"fmt"
	"os"
	"os/exec"
	"plugin/internal/commands"
	"plugin/internal/config"
	"plugin/internal/database"
	"plugin/internal/discord/bot"
	"plugin/internal/discord/webhook"
	"plugin/internal/events"
	"plugin/internal/iw4m"
	"plugin/internal/logger"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"runtime"
	"sync"

	br "plugin/internal/repository/bank"
	lr "plugin/internal/repository/link"
	pr "plugin/internal/repository/player"
	sr "plugin/internal/repository/stats"
	wr "plugin/internal/repository/wallet"

	bs "plugin/internal/service/bank"
	ls "plugin/internal/service/link"
	ps "plugin/internal/service/player"
	ss "plugin/internal/service/stats"
	ws "plugin/internal/service/wallet"
)

func ascii() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()

	// hacker ascii art type shi
	banner := []string{
		"   ____  _       _        ____  _             _       ",
		"  |  _ \\| |_   _| |_ ___ |  _ \\| |_   _  __ _(_)_ __  ",
		"  | |_) | | | | | __/ _ \\| |_) | | | | |/ _` | | '_ \\ ",
		"  |  __/| | |_| | || (_) |  __/| | |_| | (_| | | | | |",
		"  |_|   |_|\\__,_|\\__\\___/|_|   |_|\\__,_|\\__, |_|_| |_|",
		"                                        |___/         ",
	}

	lines := len(banner)
	for i, line := range banner {
		r := 180 + (100-180)*i/lines
		g := 100
		b := 255 + (100-255)*i/lines
		fmt.Printf("\033[38;2;%d;%d;%dm%s\033[0m\n", r, g, b, line)
	}

	fmt.Println("\n  Made By \033[38;2;180;100;255m@budiworld\033[0m | https://\033[38;2;180;100;255mgithub.com\033[0m/Yallamaztar/\033[38;2;180;100;255mPlutoPlugin\033[0m")
	fmt.Print("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
}
func main() {
	ascii()

	log := logger.New("PlutoPlugin", "pp_main_log.log")
	log.Println("Loading config.yaml")

	cfg, err := config.Setup(log)
	if err != nil {
		log.Fatal("config setup failed:", err)
	}

	db, err := database.Open()
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	log.Println("Starting database migrations")
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed database migration")
	}

	// creating all repositories & services
	player := ps.New(pr.New(db))
	wallet := ws.New(wr.New(db))
	bank := bs.New(br.New(db))
	link := ls.New(lr.New(db))
	playerStats := ss.NewPlayerStats(sr.NewPlayerStats(db))
	gambleStats := ss.NewGamblingStats(sr.NewGamblingStats(db))
	walletStats := ss.NewWalletStats(sr.NewWalletStats(db))
	log.Println("Database migrations done!")

	var wh *webhook.Webhook
	if cfg.Discord.Enabled {
		println()
		log.Println("Starting Discord integrations")

		if cfg.Discord.WebhookLink == "" {
			log.Fatalf("Your discord webhook link is empty, please create one and add it to the config")
		}
		wh = webhook.New(cfg.Discord.WebhookLink)

		if cfg.Discord.BotToken == "" {
			log.Fatalf("Discord bot token is empty")
		}
		discordBot, err := bot.New(cfg, log, player, wallet, walletStats, link, wh)
		if err != nil {
			log.Fatalf("Failed to create discord bot: %v\n", err)
		}

		go func() {
			if err := discordBot.Start(); err != nil {
				log.Fatalf("Failed to start discord bot: %v\n", err)
			}
		}()

		defer discordBot.Close()
		log.Infof("Successfully started discord bot")
	}

	// single IW4M-Admin wrapper instance cause the plugin only uses commands
	// like: !setlevel, !ban, !unban which are "global" commands, so handling
	// different server ids is pointless, atleast for now
	var iw *iw4m.IW4MWrapper
	if cfg.IW4MAdmin.Enabled {
		println()
		log.Println("Starting IW4M-Admin integration")
		iw = iw4m.New(cfg, log)

		err = iw.TestConnection()
		if err != nil {
			log.Fatalf(err.Error())
		}

		log.Printf("Successfully connected to IW4M-Admin (%s)\n", iw.Host)
	}

	var wg sync.WaitGroup
	for i, s := range cfg.Server {
		println()
		serverLog := logger.New(cfg.Server[i].Host, fmt.Sprintf("pp_server_log_%d.log", i))

		serverLog.Println("Creating RCON connection")
		rc, err := rcon.New(s.Host, s.Password, cfg, serverLog)
		if err != nil {
			serverLog.Fatal("Couldnt connect to RCON")
			continue
		}
		serverLog.Println("Successfully connected to RCON")

		serverLog.Println("Testing GSC connection")
		if err = rc.TestConnection(); err != nil {
			serverLog.Warnln(err)
			serverLog.Infoln("Make sure you have the necessary GSC scripts in your server /storage/t6/scripts/ dir")
			continue
		}
		serverLog.Println("Successfully verified GSC connection")

		serverLog.Println("Registering commands")
		reg := register.New(*cfg, rc, player, serverLog)
		commands.RegisterClientCommands(*cfg, rc, reg, player, wallet, bank, link, playerStats, gambleStats, walletStats, wh)

		wg.Add(1)
		go func(rc *rcon.RCON, log *logger.Logger) {
			defer wg.Done()
			defer rc.Close()
			serverLog.Println("Starting event tailer")
			events.RunEventTailLoop(i, cfg, rc, reg, iw, player, wallet, walletStats, log)
		}(rc, serverLog)
	}

	wg.Wait()
}
