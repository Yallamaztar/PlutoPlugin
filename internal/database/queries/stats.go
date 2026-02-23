package queries

const (
	CreatePlayerStats = `INSERT INTO player_stats (player_id) VALUES (?);`
	GetPlayerStats    = `SELECT player_id, total_gambles, total_wagered, total_won, total_lost, wins, losses, last_gamble FROM player_stats WHERE player_id = ?;`
	UpdateStatsWin    = `UPDATE player_stats SET total_gambles = total_gambles + 1, total_wagered = total_wagered + ?, total_won = total_won + ?, wins = wins + 1, last_gamble = CURRENT_TIMESTAMP WHERE player_id = ?;`
	UpdateStatsLoss   = `UPDATE player_stats SET total_gambles = total_gambles + 1, total_wagered = total_wagered + ?, total_lost = total_lost + ?, losses = losses + 1, last_gamble = CURRENT_TIMESTAMP WHERE player_id = ?;`
	ResetPlayerStats  = `UPDATE player_stats SET total_gambles = 0, total_wagered = 0, total_won = 0, total_lost = 0, wins = 0, losses = 0, last_gamble = NULL WHERE player_id = ?;`
)

const (
	InitGlobalStats         = `INSERT INTO global_stats (id, total_gambles, total_wagered, total_paid, last_update) VALUES (1, 0, 0, 0, CURRENT_TIMESTAMP);`
	GetGlobalStats          = `SELECT total_gambles, total_wagered, total_paid, last_update FROM global_stats WHERE id = 1;`
	UpdateGlobalAfterGamble = `UPDATE global_stats SET total_gambles = total_gambles + 1, total_wagered = total_wagered + ?, total_paid = total_paid + ?, last_update = CURRENT_TIMESTAMP WHERE id = 1;`
	ResetGlobalStats        = `UPDATE global_stats SET total_gambles = 0, total_wagered = 0, total_paid = 0, last_update = CURRENT_TIMESTAMP WHERE id = 1;`
)

const (
	CreateWalletStats = `INSERT INTO wallet_stats (player_id, balance, total_paid, total_received, deposit_count, withdraw_count) VALUES (?, 0, 0, 0, 0, 0);`
	GetWalletStats    = `SELECT player_id, balance, total_paid, total_received, deposit_count, withdraw_count FROM wallet_stats WHERE player_id = ?;`
	WalletDeposit     = `UPDATE wallet_stats SET balance = balance + ?, total_received = total_received + ?, deposit_count = deposit_count + 1 WHERE player_id = ?;`
	WalletWithdraw    = `UPDATE wallet_stats SET balance = balance - ?, total_paid = total_paid + ?, withdraw_count = withdraw_count + 1 WHERE player_id = ?;`
	WalletPay         = `UPDATE wallet_stats SET balance = balance + ?, total_paid = total_paid + ? WHERE player_id = ?;`
	WalletReceive     = `UPDATE wallet_stats SET balance = balance - ?, total_received = total_received + ? WHERE player_id = ?;`
	ResetWalletStats  = `UPDATE wallet_stats SET balance = 0, total_paid = 0, total_received = 0, deposit_count = 0, withdraw_count = 0 WHERE player_id = ?;`
)
