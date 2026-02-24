package main

import (
	"fmt"
	"plugin/internal/commands"
	"plugin/internal/config"
	"plugin/internal/database"
	"plugin/internal/events"
	"plugin/internal/iw4m"
	"plugin/internal/logger"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"sync"

	br "plugin/internal/repository/bank"
	pr "plugin/internal/repository/player"
	sr "plugin/internal/repository/stats"
	wr "plugin/internal/repository/wallet"

	bs "plugin/internal/service/bank"
	ps "plugin/internal/service/player"
	ss "plugin/internal/service/stats"
	ws "plugin/internal/service/wallet"
)

func main() {
	log := logger.New("main", "pp_main_log.log")

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

	player := ps.New(pr.New(db))
	wallet := ws.New(wr.New(db))
	bank := bs.New(br.New(db))
	playerStats := ss.NewPlayerStats(sr.NewPlayerStats(db))
	gambleStats := ss.NewGamblingStats(sr.NewGamblingStats(db))
	walletStats := ss.NewWalletStats(sr.NewWalletStats(db))
	log.Println("Database migrations done!")

	var iw *iw4m.IW4MWrapper
	if cfg.IW4MAdmin.Enabled {
		log.Println("Starting IW4M-Admin integration")
		iw = iw4m.New(cfg)
	}

	var wg sync.WaitGroup
	for i, s := range cfg.Server {
		serverLog := logger.New(cfg.Server[i].Host, fmt.Sprintf("pp_server_log_%d.log", i))
		serverLog.Println("Connecting RCON")

		rc, err := rcon.New(s.Host, s.Password, cfg, serverLog)
		if err != nil {
			serverLog.Println("Couldnt connect to RCON")
			continue
		}

		if err = rc.TestConnection(); err != nil {
			serverLog.Println(err)
			continue
		}

		reg := register.New(*cfg, rc, player)
		commands.RegisterClientCommands(*cfg, rc, reg, player, wallet, bank, playerStats, gambleStats, walletStats)

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
