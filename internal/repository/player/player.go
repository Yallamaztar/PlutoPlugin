package player

import (
	"database/sql"
	"errors"
	"plugin/internal/database/queries"
	"time"
)

type Player struct {
	ID        int
	Name      string
	XUID      string
	GUID      string
	Level     int
	ClientID  *int
	DiscordID *string
	CreatedAt time.Time
}

type PlayerRepository interface {
	Create(name, xuid, guid string, level int, clientID *int) (int, error)
	GetByID(id int) (*Player, error)
	GetByXUID(xuid string) (*Player, error)
	GetByGUID(guid string) (*Player, error)
	GetDiscordIDByID(id int) (string, error)
	GetPlayerByDiscordID(id string) (*Player, error)
	UpdateDiscordID(id int, discordID string) error
	UpdateName(id int, name string) error
	UpdateLevel(id int, level int) error
	Delete(id int) error
	ExistsByID(id int) (bool, error)
	ExistsByXUID(xuid string) (bool, error)
	ExistsByGUID(guid string) (bool, error)
	GetAll() ([]*Player, error)
}

type repository struct {
	db *sql.DB
}

func New(db *sql.DB) PlayerRepository {
	return &repository{db: db}
}

func (r *repository) Create(name, xuid, guid string, level int, clientID *int) (int, error) {
	res, err := r.db.Exec(queries.CreatePlayer, name, xuid, guid, level, clientID)
	if err != nil {
		return 0, err
	}
	n, err := res.LastInsertId()
	return int(n), err
}

func (r *repository) GetByID(id int) (*Player, error) {
	var p Player
	err := r.db.QueryRow(queries.GetPlayerByID, id).Scan(
		&p.ID, &p.Name, &p.XUID, &p.GUID, &p.Level, &p.ClientID, &p.DiscordID, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *repository) GetByXUID(xuid string) (*Player, error) {
	var p Player
	err := r.db.QueryRow(queries.GetPlayerByXUID, xuid).Scan(
		&p.ID, &p.Name, &p.XUID, &p.GUID, &p.Level, &p.ClientID, &p.DiscordID, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *repository) GetByGUID(guid string) (*Player, error) {
	var p Player
	err := r.db.QueryRow(queries.GetPlayerByGUID, guid).Scan(
		&p.ID, &p.Name, &p.XUID, &p.GUID, &p.Level, &p.ClientID, &p.DiscordID, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *repository) GetDiscordIDByID(id int) (string, error) {
	var discorID string
	err := r.db.QueryRow(queries.GetDiscordIDByID, id).Scan(&discorID)
	if err != nil {
		return "", err
	}
	return discorID, nil
}
func (r *repository) GetDiscordIDByXUID(xuid string) (string, error) {
	var discorID string
	err := r.db.QueryRow(queries.GetDiscordIDByID, xuid).Scan(&discorID)
	if err != nil {
		return "", err
	}
	return discorID, nil
}

func (r *repository) GetPlayerByDiscordID(id string) (*Player, error) {
	var p Player
	err := r.db.QueryRow(queries.GetPlayerByDiscordID, id).Scan(
		&p.ID, &p.Name, &p.XUID, &p.GUID, &p.Level, &p.ClientID, &p.DiscordID, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *repository) UpdateDiscordID(id int, discordID string) error {
	_, err := r.db.Exec(queries.UpdatePlayerDiscordID, discordID, id)
	return err
}

func (r *repository) UpdateName(id int, name string) error {
	_, err := r.db.Exec(queries.UpdatePlayerName, name, id)
	return err
}

func (r *repository) UpdateLevel(id int, level int) error {
	_, err := r.db.Exec(queries.UpdatePlayerLevel, level, id)
	return err
}

func (r *repository) Delete(id int) error {
	_, err := r.db.Exec(queries.DeletePlayerByID, id)
	return err
}

func (r *repository) ExistsByID(id int) (bool, error) {
	var count int
	err := r.db.QueryRow(queries.UserExistsByID, id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) ExistsByXUID(xuid string) (bool, error) {
	var count int
	err := r.db.QueryRow(queries.UserExistsByXUID, xuid).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) ExistsByGUID(guid string) (bool, error) {
	var count int
	err := r.db.QueryRow(queries.UserExistsByGUID, guid).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) GetAll() ([]*Player, error) {
	rows, err := r.db.Query(queries.GetAllUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []*Player
	for rows.Next() {
		var p Player
		if err := rows.Scan(&p.ID, &p.Name, &p.XUID, &p.GUID, &p.Level, &p.CreatedAt); err != nil {
			return nil, err
		}
		players = append(players, &p)
	}
	return players, nil
}
