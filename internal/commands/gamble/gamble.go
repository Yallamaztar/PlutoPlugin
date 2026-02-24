package gamble

import (
	"errors"
	"fmt"
	"math/rand"
	"plugin/internal/config"
	"plugin/internal/service/bank"
	"plugin/internal/service/player"
	"plugin/internal/service/stats"
	"plugin/internal/service/wallet"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

type Result struct {
	Won     bool
	Amount  int
	Balance int
	Message string
}

func Gamble(
	playerID, amount int,
	cfg config.Config,
	player *player.Service,
	wallet *wallet.Service,
	bank *bank.Service,

	playerStats *stats.PlayeStatsService,
	gambleStats *stats.GamblingStatsService,
	walletStats *stats.WalletStatsService,
) (*Result, error) {
	if amount <= 0 {
		return nil, errors.New("invalid gamble amount")
	}

	bet := int(amount)

	balance, err := wallet.GetBalance(playerID)
	if err != nil {
		return nil, err
	}

	if balance < bet {
		return nil, errors.New("You dont have enough money")
	}

	if didWin(cfg.Gambling.WinChance) {
		if err := bank.Withdraw(bet); err != nil {
			return nil, err
		}

		if err := wallet.Deposit(playerID, bet); err != nil {
			_ = bank.Deposit(bet)
			return nil, err
		}

		if err := playerStats.Win(playerID, bet, bet*2); err != nil {
			return nil, err
		}

		if err := gambleStats.RecordGamble(bet, bet*2); err != nil {
			return nil, err
		}

		if err := walletStats.Deposit(playerID, bet); err != nil {
			return nil, err
		}

		return &Result{
			Won:     true,
			Amount:  bet,
			Balance: balance + bet,
			Message: fmt.Sprintf("You just won %s%d!", cfg.Gambling.Currency, bet),
		}, nil
	}

	if err := wallet.Withdraw(playerID, bet); err != nil {
		return nil, err
	}

	if err := bank.Deposit(bet); err != nil {
		return nil, err
	}

	if err := playerStats.Loss(playerID, bet); err != nil {
		return nil, err
	}

	if err := gambleStats.RecordGamble(bet, 0); err != nil {
		return nil, err
	}

	if err := walletStats.Withdraw(playerID, bet); err != nil {
		return nil, err
	}

	return &Result{
		Won:     false,
		Amount:  bet,
		Balance: balance - bet,
		Message: fmt.Sprintf("You just lost %s%d!", cfg.Gambling.Currency, bet),
	}, nil
}

func didWin(winChance float64) bool {
	r1 := rng.Float64()
	r2 := rng.Float64()
	r3 := rng.Float64()

	avg := (r1 + r2 + r3) / 3
	return avg < winChance
}
