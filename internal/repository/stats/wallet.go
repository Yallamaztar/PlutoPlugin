package stats

import (
	"database/sql"
	"errors"
	"plugin/internal/database/queries"
)

type WalletStats struct {
	PlayerID int

	Balance       int
	TotalPaid     int
	TotalReceived int

	DepositCount  int
	WithdrawCount int

	PlayerStats
}

type WalletStatsRepository interface {
	Create(playerID int) error
	Get(playerID int) (*WalletStats, error)

	Deposit(playerID int, amount int) error
	Withdraw(playerID int, amount int) error

	Pay(playerID int, amount int) error     // payout (e.g. winnings)
	Receive(playerID int, amount int) error // wager or other incoming funds

	Reset(playerID int) error
}

type walletRepository struct {
	db *sql.DB
}

func NewWalletStats(db *sql.DB) WalletStatsRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(playerID int) error {
	_, err := r.db.Exec(queries.CreateWalletStats, playerID)
	return err
}

func (r *walletRepository) Get(playerID int) (*WalletStats, error) {
	var stats WalletStats

	err := r.db.QueryRow(queries.GetWalletStats, playerID).Scan(
		&stats.PlayerID,
		&stats.Balance,
		&stats.TotalPaid,
		&stats.TotalReceived,
		&stats.DepositCount,
		&stats.WithdrawCount,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &stats, nil
}

func (r *walletRepository) Deposit(playerID int, amount int) error {
	_, err := r.db.Exec(queries.WalletDeposit, amount, playerID)
	return err
}

func (r *walletRepository) Withdraw(playerID int, amount int) error {
	_, err := r.db.Exec(queries.WalletWithdraw, amount, playerID)
	return err
}

func (r *walletRepository) Pay(playerID int, amount int) error {
	_, err := r.db.Exec(queries.WalletPay, amount, playerID)
	return err
}

func (r *walletRepository) Receive(playerID int, amount int) error {
	_, err := r.db.Exec(queries.WalletReceive, amount, playerID)
	return err
}

func (r *walletRepository) Reset(playerID int) error {
	_, err := r.db.Exec(queries.ResetWalletStats, playerID)
	return err
}
