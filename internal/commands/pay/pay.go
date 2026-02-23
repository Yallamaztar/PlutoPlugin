package pay

import (
	"errors"
	"fmt"
	"plugin/internal/config"
	"plugin/internal/service/player"
	"plugin/internal/service/stats"
	"plugin/internal/service/wallet"
)

type Result struct {
	FromPlayer  int
	ToPlayer    int
	Amount      int
	FromBalance int
	ToBalance   int
	Message     string
}

func Pay(
	fromPlayerID int,
	toPlayerID int,
	amount int,

	cfg config.Config,
	player *player.Service,
	wallet *wallet.Service,
	walletStats *stats.WalletStatsService,

) (*Result, error) {
	if amount <= 0 {
		return nil, errors.New("invalid amount")
	}

	if _, err := player.GetPlayerByID(fromPlayerID); err != nil {
		return nil, errors.New("error occurred, please try again later")
	}

	if _, err := player.GetPlayerByID(toPlayerID); err != nil {
		return nil, fmt.Errorf("receiver (%d) doesnt exists", toPlayerID)
	}

	fromBalance, err := wallet.GetBalance(fromPlayerID)
	if err != nil {
		return nil, err
	}

	if fromBalance < amount {
		return nil, errors.New("You dont have enough money")
	}

	if err := wallet.Withdraw(fromPlayerID, amount); err != nil {
		return nil, err
	}

	if err := wallet.Deposit(toPlayerID, amount); err != nil {
		_ = wallet.Deposit(fromPlayerID, amount)
		return nil, err
	}

	toBalance, err := wallet.GetBalance(toPlayerID)
	if err != nil {
		return nil, err
	}

	walletStats.Pay(fromPlayerID, amount)
	walletStats.Receive(toPlayerID, amount)

	return &Result{
		FromPlayer:  fromPlayerID,
		ToPlayer:    toPlayerID,
		Amount:      amount,
		FromBalance: fromBalance - amount,
		ToBalance:   toBalance,
		Message: fmt.Sprintf(
			"You paid %s%d to player %d",
			cfg.Gambling.Currency,
			amount,
			toPlayerID,
		),
	}, nil
}
