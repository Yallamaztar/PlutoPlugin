package events

import (
	"context"
	"fmt"
	"plugin/internal/config"
	"plugin/internal/logger"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"strings"

	ps "plugin/internal/service/player"
	ss "plugin/internal/service/stats"
	ws "plugin/internal/service/wallet"

	"github.com/Yallamaztar/eventsv2/events"
)

func RunEventTailLoop(
	index int,
	cfg *config.Config,
	rcon *rcon.RCON,
	reg *register.Register,

	playerService *ps.Service,
	walletService *ws.Service,
	walletStats *ss.WalletStatsService,

	log *logger.Logger,
) {
	eventsCh := make(chan events.Event, 128)
	server := cfg.Server[index]
	go func(cfg *config.Config) {
		if err := events.TailFileContext(context.Background(), server.LogPath, false, eventsCh); err != nil {
			log.Fatalf("Failed to tail file: %s", server)
			return
		}
		close(eventsCh)
	}(cfg)

	for e := range eventsCh {
		if !cfg.Gambling.Enabled {
			continue
		}

		switch event := e.(type) {
		case *events.PlayerEvent:
			switch event.Command {
			case "J":
				go func() {
					reg.SetClientNum(event.XUID, event.ClientNum)

					guid := rcon.GUIDByClientNum(event.ClientNum)
					if guid == "" {
						return
					}

					exists, err := playerService.ExistsByXUID(event.XUID)
					if err != nil {
						return
					}

					if exists {
						p, err := playerService.GetPlayerByXUID(event.XUID)
						if err != nil {
							return
						}

						walletService.Deposit(p.ID, int(cfg.Economy.JoinReward))
						walletStats.Deposit(p.ID, int(cfg.Economy.JoinReward))
						return
					}

					id, err := playerService.CreatePlayer(event.Name, event.XUID, guid, 0)
					if err != nil {
						return
					}

					err = walletService.CreateWallet(id, int(cfg.Economy.FirstTimeReward))
					if err != nil {
						playerService.DeletePlayer(id)
						return
					}

					walletStats.Init(id)
					walletStats.Deposit(id, int(cfg.Economy.FirstTimeReward))

					log.Printf("Created wallet: %s (%s) | ID: %d\n", event.Name, event.XUID, id)
					rcon.Tell(
						event.ClientNum,
						fmt.Sprintf(
							"^7Created a wallet with ^6%s%d balance",
							cfg.Gambling.Currency,
							cfg.Economy.FirstTimeReward,
						),
					)
				}()
				continue

			case "Q":
				go reg.RemoveClientNum(event.XUID)

			case "say", "sayteam":
				if cmd, isCommand := strings.CutPrefix(event.Message, server.CommandPrefix); isCommand {
					parts := strings.Fields(cmd)
					if len(parts) > 0 {
						args := []string{}
						if len(parts) > 1 {
							args = parts[1:]
						}

						p, err := playerService.GetPlayerByXUID(event.XUID)
						if err != nil || p == nil {
							continue
						}

						go reg.Execute(event.ClientNum, p.ID, event.Name, event.XUID, p.Level, parts[0], args)
					}
				}
			}
		}
	}
}
