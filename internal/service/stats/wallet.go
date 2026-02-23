package stats

import (
	"fmt"
	repo "plugin/internal/repository/stats"
)

type WalletStatsService struct {
	repo repo.WalletStatsRepository
}

func NewWalletStats(repo repo.WalletStatsRepository) *WalletStatsService {
	return &WalletStatsService{repo: repo}
}

func (s *WalletStatsService) Init(playerID int) error {
	stats, err := s.repo.Get(playerID)
	if err != nil {
		return err
	}
	if stats != nil {
		return nil
	}
	return s.repo.Create(playerID)
}

func (s *WalletStatsService) GetStats(playerID int) (*repo.WalletStats, error) {
	return s.repo.Get(playerID)
}

func (s *WalletStatsService) Deposit(playerID int, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("invalid deposit amount")
	}
	return s.repo.Deposit(playerID, amount)
}

func (s *WalletStatsService) Withdraw(playerID int, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("invalid withdraw amount")
	}
	return s.repo.Withdraw(playerID, amount)
}

func (s *WalletStatsService) Pay(playerID int, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("invalid payout amount")
	}
	return s.repo.Pay(playerID, amount)
}

func (s *WalletStatsService) Receive(playerID int, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("invalid received amount")
	}
	return s.repo.Receive(playerID, amount)
}

func (s *WalletStatsService) Reset(playerID int) error {
	return s.repo.Reset(playerID)
}
