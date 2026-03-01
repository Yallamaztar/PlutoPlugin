package gamble

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"plugin/internal/config"
	"plugin/internal/discord/webhook"
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
	playerID int,
	playerName string,
	amount int,

	cfg *config.Config,
	player *player.Service,
	wallet *wallet.Service,
	bank *bank.Service,

	playerStats *stats.PlayeStatsService,
	gambleStats *stats.GamblingStatsService,
	walletStats *stats.WalletStatsService,

	webhook *webhook.Webhook,
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

		if cfg.Discord.Enabled {
			webhook.WinWebhook(playerName, bet)
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

	if cfg.Discord.Enabled {
		webhook.LossWebhook(playerName, bet)
	}

	return &Result{
		Won:     false,
		Amount:  bet,
		Balance: balance - bet,
		Message: fmt.Sprintf("You just lost %s%d!", cfg.Gambling.Currency, bet),
	}, nil
}

// scizophrenic maniac paranoid level of randomness
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))
var paranoia uint64 = uint64(time.Now().UnixNano())

func xorshift64(x uint64) uint64 {
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	return x
}

func mix64(x uint64) uint64 {
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return x
}

func paranoidFloat() float64 {
	r := rng.Uint64()
	t := uint64(time.Now().UnixNano())
	paranoia = xorshift64(paranoia + t + r)
	mixed := mix64(r ^ paranoia ^ t)
	return float64(mixed>>11) * (1.0 / (1 << 53))
}

func didWin(winChance float64) bool {
	v := paranoidFloat()
	v = math.Mod(math.Sin(v*1e6+math.Phi)*1e5, 1.0)
	if v < 0 {
		v += 1
	}

	return v < winChance
}
