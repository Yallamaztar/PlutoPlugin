package commands

import (
	"fmt"
	"plugin/internal/commands/gamble"
	"plugin/internal/commands/pay"
	"plugin/internal/config"
	"plugin/internal/helpers"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"plugin/internal/service/bank"
	"plugin/internal/service/player"
	"plugin/internal/service/stats"
	"plugin/internal/service/wallet"
)

const (
	LevelUser = iota
	levelAdmin
	LevelOwner
)

const unknownErr = "an ^1error ^7occurred, please ^1try again ^7later"

func RegisterClientCommands(
	cfg config.Config,

	rcon *rcon.RCON,
	reg *register.Register,

	player *player.Service,
	wallet *wallet.Service,
	bank *bank.Service,

	playerStats *stats.PlayeStatsService,
	gambleStats *stats.GamblingStatsService,
	walletStats *stats.WalletStatsService,

) {
	reg.RegisterCommand(register.Command{
		Name:     "gamble",
		Aliases:  []string{"g", "cf", "coinflip"},
		MinLevel: LevelUser,
		Help:     "Usage: ^6!gamble ^7<amount>",
		MinArgs:  1,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			balance, err := wallet.GetBalance(playerID)
			if err != nil {
				rcon.Tell(clientNum, unknownErr)
				return
			}

			amount, err := helpers.ParseAmountArg(args[0], int64(balance))
			if err != nil {
				rcon.Tell(clientNum, fmt.Sprintf("%s (%q)", err, args[0]))
				return
			}

			res, err := gamble.Gamble(playerID, amount, cfg, player, wallet, bank, playerStats, gambleStats, walletStats)
			if err != nil {
				rcon.Tell(clientNum, err.Error())
				return
			}

			rcon.Tell(clientNum, res.Message)
			rcon.Say(fmt.Sprintf("%s just ^6won ^7%s%d!", playerName, cfg.Gambling.Currency, res.Amount))
		},
	})

	reg.RegisterCommand(register.Command{
		Name:     "pay",
		Aliases:  []string{"pp", "payplayer", "transfer"},
		MinLevel: LevelUser,
		Help:     "Usage: ^6!pay <player> <amount>",
		MinArgs:  2,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			balance, err := wallet.GetBalance(playerID)
			if err != nil {
				rcon.Tell(clientNum, unknownErr)
				return
			}

			amount, err := helpers.ParseAmountArg(args[0], int64(balance))
			if err != nil {
				rcon.Tell(clientNum, fmt.Sprintf("%s (%q)", err, args[0]))
				return
			}

			t := reg.FindPlayer(args[0])
			if t == nil {
				rcon.Tell(clientNum, fmt.Sprintf("player %s couldnt be found", args[0]))
				return
			}

			target, err := player.GetPlayerByGUID(t.GUID)
			if err != nil {
				rcon.Tell(clientNum, unknownErr)
				return
			}

			res, err := pay.Pay(playerID, target.ID, amount, cfg, player, wallet, walletStats)
			if err != nil {
				rcon.Tell(clientNum, err.Error())
				return
			}

			rcon.Tell(clientNum, res.Message)
		},
	})

	reg.RegisterCommand(register.Command{
		Name:     "balance",
		Aliases:  []string{"bal", "balanc", "money"},
		MinLevel: LevelUser,
		Help:     "Usage: ^6!balance <player>",
		MinArgs:  1,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			bal, err := wallet.GetBalance(playerID)
			if err != nil {
				rcon.Tell(clientNum, unknownErr)
				return
			}

			rcon.Tell(clientNum, fmt.Sprintf("Your balance is %s%d", cfg.Gambling.Currency, bal))
		},
	})

	reg.RegisterCommand(register.Command{
		Name:     "bankbalance",
		Aliases:  []string{"bb", "bank", "bankbal"},
		MinLevel: LevelUser,
		Help:     "Usage: ^6!bankbalance",
		MinArgs:  1,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			bal, err := bank.GetBalance()
			if err != nil {
				rcon.Tell(clientNum, unknownErr)
				return
			}

			rcon.Tell(clientNum, fmt.Sprintf("bank balance is ^6%s%d", cfg.Gambling.Currency, bal))
		},
	})

	reg.RegisterCommand(register.Command{
		Name:     "discord",
		Aliases:  []string{"dc", "disc"},
		MinLevel: LevelUser,
		Help:     "Usage: ^6!discord",
		MinArgs:  1,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			rcon.Tell(clientNum, cfg.Discord.InviteLink)
		},
	})

	reg.RegisterCommand(register.Command{
		Name:     "richest",
		Aliases:  []string{"rich"},
		MinLevel: LevelUser,
		Help:     "Usage: ^6!richest",
		MinArgs:  1,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {

		},
	})

	reg.RegisterCommand(register.Command{
		Name:     "poorest",
		Aliases:  []string{"poor"},
		MinLevel: LevelUser,
		Help:     "Usage: ^6!poorest",
		MinArgs:  1,
		Handler:  func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {},
	})

}
