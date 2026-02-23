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

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	won := rng.Float64() < cfg.Gambling.WinChance

	if won {
		if err := bank.Withdraw(bet); err != nil {
			return nil, err
		}

		if err := wallet.Deposit(playerID, bet); err != nil {
			_ = bank.Deposit(bet)
			return nil, err
		}

		_ = playerStats.Win(playerID, bet, bet*2)
		_ = gambleStats.RecordGamble(bet, bet*2)
		_ = walletStats.Deposit(playerID, bet)

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

	_ = playerStats.Loss(playerID, bet)
	_ = gambleStats.RecordGamble(bet, 0)
	_ = walletStats.Withdraw(playerID, bet)

	return &Result{
		Won:     false,
		Amount:  bet,
		Balance: balance - bet,
		Message: fmt.Sprintf("You just lost %s%d!", cfg.Gambling.Currency, bet),
	}, nil
}
