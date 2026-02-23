package queries

const (
	CreatePlayer     = `INSERT INTO players (name, xuid, guid, level, client_id, created_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP);`
	DeletePlayerByID = `DELETE FROM players WHERE id = ?;`

	GetPlayerByID   = `SELECT id, name, xuid, guid, level, client_id, created_at FROM players WHERE id = ?;`
	GetPlayerByXUID = `SELECT id, name, xuid, guid, level, client_id, created_at FROM players WHERE xuid = ?;`
	GetPlayerByGUID = `SELECT id, name, xuid, guid, level, client_id, created_at FROM players WHERE guid = ?;`

	GetIDByXUID = `SELECT id FROM players WHERE xuid = ?;`
	GetIDByGUID = `SELECT id FROM players WHERE guid = ?;`

	UpdatePlayerLevel = `UPDATE players SET level = ? WHERE id = ?;`
	UpdatePlayerName  = `UPDATE players SET name = ? WHERE id = ?;`

	UserExistsByID   = `SELECT COUNT(1) FROM players WHERE id = ?;`
	UserExistsByXUID = `SELECT COUNT(1) FROM players WHERE xuid = ?;`
	UserExistsByGUID = `SELECT COUNT(1) FROM players WHERE guid = ?;`

	GetAllUsers = `SELECT id, name, xuid, guid, level, client_id, created_at FROM players ORDER BY id ASC;`
)
