package events

import (
	"context"
	"fmt"
	"math/rand"
	"plugin/internal/config"
	"plugin/internal/iw4m"
	"plugin/internal/logger"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"strings"
	"sync"
	"time"

	ps "plugin/internal/service/player"
	ss "plugin/internal/service/stats"
	ws "plugin/internal/service/wallet"

	"github.com/Yallamaztar/eventsv2/events"
)

var (
	sessionStats = make(map[string]*session)
	mostvaluable = make(map[string]int)

	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type session struct {
	Kills  int
	Deaths int
	mu     sync.Mutex
}

func RunEventTailLoop(
	index int,
	cfg *config.Config,
	rcon *rcon.RCON,
	reg *register.Register,
	iw4m *iw4m.IW4MWrapper,

	playerService *ps.Service,
	walletService *ws.Service,
	walletStats *ss.WalletStatsService,

	log *logger.Logger,
) {
	eventsCh := make(chan events.Event, 128)
	server := cfg.Server[index]
	go func(cfg *config.Config) {
		if err := events.TailFileContext(context.Background(), server.LogPath, true, eventsCh); err != nil {
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

					if !exists {
						id, err := playerService.CreatePlayer(event.Name, event.XUID, guid, 0, cfg, iw4m)
						if err != nil {
							return
						}

						err = walletService.CreateWallet(id, cfg.Economy.FirstTimeReward)
						if err != nil {
							playerService.DeletePlayer(id)
							return
						}

						walletStats.Init(id)
						walletStats.Deposit(id, cfg.Economy.FirstTimeReward)

						log.Printf("Created wallet: %s (%s) | ID: %d\n", event.Name, event.XUID, id)
						rcon.Tell(
							event.ClientNum,
							fmt.Sprintf(
								"^7Created a wallet with ^6%s%d balance",
								cfg.Gambling.Currency,
								cfg.Economy.FirstTimeReward,
							),
						)
					}

					p, err := playerService.GetPlayerByXUID(event.XUID)
					if err != nil {
						return
					}

					if cfg.IW4MAdmin.Enabled {
						stats, err := iw4m.Stats(*p.ClientID, index)
						if err != nil {
							return
						}

						reward := calcJoinReward(stats.TotalSecondsPlayed, stats.Kills, stats.Deaths, cfg.Economy.JoinReward)
						walletService.Deposit(p.ID, reward)
						walletStats.Deposit(p.ID, reward)
						rcon.Tell(event.ClientNum, fmt.Sprintf("^7Spawning bonus: %s%d", cfg.Gambling.Currency, reward))
					} else {
						walletService.Deposit(p.ID, cfg.Economy.JoinReward)
						walletStats.Deposit(p.ID, cfg.Economy.JoinReward)
						rcon.Tell(event.ClientNum, fmt.Sprintf("^7Spawning bonus: %s%d", cfg.Gambling.Currency, cfg.Economy.JoinReward))
					}
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
		case *events.KillEvent:
			go func() {
				handleKillEvent(event.AttackerXUID)
				handleDeathEvent(event.VictimXUID)

				mvp, ok := mostvaluable[event.VictimXUID]
				if !ok {
					return
				}
				rcon.Tell(event.AttackerClientNum, fmt.Sprintf("Claimed bounty: %s%d for killing %s", cfg.Gambling.Currency, mvp, event.VictimName))
				rcon.Say(fmt.Sprintf("%s has killed MVP %s and got rewarded: %s%d", event.AttackerName, event.VictimName, cfg.Gambling.Currency, mvp))

				mostvaluable = make(map[string]int)
			}()

		case *events.ServerEvent:
			if event.Command == "InitGame" {
				go func() {
					status, err := rcon.Status()
					if err != nil {
						return
					}

					if len(status.Players) < 4 {
						return
					}

					var (
						topKills int
						topXUID  string
						topName  string
					)

					for _, p := range status.Players {
						p, err := playerService.GetPlayerByGUID(p.GUID)
						if err != nil {
							return
						}

						s, ok := sessionStats[p.XUID]
						if !ok {
							continue
						}

						s.mu.Lock()
						kills := s.Kills
						s.mu.Unlock()

						if kills > topKills {
							topKills = kills
							topXUID = p.XUID
							topName = p.Name
						}

					}

					if topKills < 7 {
						return
					}

					reward := randomReward()
					mostvaluable[topXUID] = reward

					rcon.Say(fmt.Sprintf(
						"^6BOUNTY ACTIVE! ^7Kill %s to claim the bounty!",
						topName,
					))
				}()
			}

			if event.Command == "ShutdownGame" {
				go func() {
					status, err := rcon.Status()
					if err != nil {
						return
					}

					if len(status.Players) < 4 {
						return
					}

					for _, ps := range status.Players {
						p, err := playerService.GetPlayerByGUID(ps.GUID)
						if err != nil {
							return
						}

						s, ok := mostvaluable[p.XUID]
						if !ok {
							continue
						}

						cn := rcon.ClientNumByGUID(ps.GUID)
						if cn == -1 {
							break
						}

						reward := int(s / 2)

						walletService.Deposit(p.ID, reward)
						rcon.Tell(uint8(cn), fmt.Sprintf("You survived the round as MVP and got rewarded: %s%d", cfg.Gambling.Currency, reward))
					}
				}()
			}
		}
	}
}

func randomReward() int {
	v := rng.Float64()
	return int((v*v)*199500) + 500
}

func handleKillEvent(xuid string) {
	stats, ok := sessionStats[xuid]
	if !ok {
		stats = &session{mu: sync.Mutex{}}
		sessionStats[xuid] = stats
	}
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.Kills++
}

func handleDeathEvent(xuid string) {
	stats, ok := sessionStats[xuid]
	if !ok {
		stats = &session{mu: sync.Mutex{}}
		sessionStats[xuid] = stats
	}
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.Deaths++
}

func calcJoinReward(seconds, kills, deaths, joinReward int) int {
	var reward int
	switch {
	case seconds <= 18000:
		reward = joinReward
	case seconds <= 72000:
		reward = joinReward * 3
	case seconds <= 180000:
		reward = joinReward * 5
	default:
		reward = joinReward * 8
	}

	kdr := calcKDR(kills, deaths)
	switch {
	case kdr >= 3.0:
		reward += joinReward * 4
	case kdr >= 2.0:
		reward += joinReward * 3
	case kdr >= 1.5:
		reward += joinReward * 2
	case kdr >= 1.0:
		reward += joinReward
	default:
	}

	return reward
}

func calcKDR(kills, deaths int) float64 {
	if deaths == 0 {
		if kills == 0 {
			return 0
		}
		return float64(kills)
	}
	return float64(kills) / float64(deaths)
}
