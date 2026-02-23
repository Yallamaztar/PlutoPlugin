package player

import (
	"fmt"
	"plugin/internal/config"
	"plugin/internal/iw4m"
	"plugin/internal/repository/player"
)

type Service struct {
	repo player.PlayerRepository
}

func New(repo player.PlayerRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreatePlayer(name, xuid, guid string, level int, cfg *config.Config, iw4m *iw4m.IW4MWrapper) (int, error) {
	existsXUID, err := s.repo.ExistsByXUID(xuid)
	if err != nil {
		return 0, err
	}
	if existsXUID {
		return 0, fmt.Errorf("player with XUID %s already exists", xuid)
	}

	existsGUID, err := s.repo.ExistsByGUID(guid)
	if err != nil {
		return 0, err
	}
	if existsGUID {
		return 0, fmt.Errorf("player with GUID %s already exists", guid)
	}

	var clientID *int
	if cfg.IW4MAdmin.Enabled {
		clientID = iw4m.ClientIDFromGUID(guid)
	}
	return s.repo.Create(name, xuid, guid, level, clientID)
}

func (s *Service) GetPlayerByID(id int) (*player.Player, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetPlayerByXUID(xuid string) (*player.Player, error) {
	return s.repo.GetByXUID(xuid)
}

func (s *Service) GetPlayerByGUID(guid string) (*player.Player, error) {
	return s.repo.GetByGUID(guid)
}

func (s *Service) UpdateName(id int, name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return s.repo.UpdateName(id, name)
}

func (s *Service) UpdateLevel(id int, level int) error {
	if level < 0 {
		return fmt.Errorf("level cannot be negative")
	}
	return s.repo.UpdateLevel(id, level)
}

func (s *Service) DeletePlayer(id int) error {
	return s.repo.Delete(id)
}

func (s *Service) ExistsByID(id int) (bool, error) {
	return s.repo.ExistsByID(id)
}

func (s *Service) ExistsByXUID(xuid string) (bool, error) {
	return s.repo.ExistsByXUID(xuid)
}

func (s *Service) ExistsByGUID(guid string) (bool, error) {
	return s.repo.ExistsByGUID(guid)
}

func (s *Service) GetAllPlayers() ([]*player.Player, error) {
	return s.repo.GetAll()
}
