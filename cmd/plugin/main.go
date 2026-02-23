package main

import (
	"plugin/internal/commands"
	"plugin/internal/config"
	"plugin/internal/database"
	"plugin/internal/events"
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
	log := logger.New("main")

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
	log.Println("Database migrations done!")

	var wg sync.WaitGroup
	for i, s := range cfg.Server {
		serverLog := logger.New(cfg.Server[i].Host)
		serverLog.Println("Connecting to RCON on " + s.Host)
		rc, err := rcon.New(s.Host, s.Password, cfg)
		if err != nil {
			serverLog.Println("Couldnt connect to RCON on " + s.Host)
			continue
		}

		reg := register.New(*cfg, rc, player)
		commands.RegisterClientCommands(*cfg, rc, reg, player, wallet, bank, playerStats, gambleStats)

		serverLog.Println("Starting Plugin")
		wg.Add(1)
		go func(rc *rcon.RCON, log *logger.Logger) {
			defer wg.Done()
			defer rc.Close()

			events.RunEventTailLoop(i, cfg, rc, reg, player, wallet, log)
		}(rc, serverLog)
	}

	wg.Wait()

}
