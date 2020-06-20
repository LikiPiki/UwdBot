package database

import (
	"context"
)

type User struct {
	ID         uint64
	UserID     uint64
	Username   string
	Coins      int
	Reputation int
	Blacklist  bool
	IsAdmin    bool
}

func (u *User) CreateNewUser() (uint64, error) {
	row := db.QueryRow(
		context.Background(),
		"INSERT INTO users (username, userID, coins) VALUES ($1, $2, 100) RETURNING id",
		u.Username,
		u.UserID,
	)
	err := row.Scan(&u.ID)
	if err != nil {
		return 0, err
	}

	return u.ID, nil
}

func (u *User) CountUsersWithID(id int) (int, error) {
	var count int
	row := db.QueryRow(
		context.Background(),
		"SELECT COUNT (*) FROM users WHERE userID = $1",
		id,
	)

	err := row.Scan(
		&count,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (u *User) DeleteUser(id int) (int, error) {
	var count int
	row := db.QueryRow(
		context.Background(),
		"DELETE FROM users WHERE userID = $1",
		id,
	)

	err := row.Scan(
		&count,
	)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (u *User) FindUserByID(id int) (User, error) {
	row := db.QueryRow(
		context.Background(),
		"SELECT id, username, userid, blacklist, isadmin, coins, reputation FROM users WHERE userID = $1",
		id,
	)
	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.UserID,
		&u.Blacklist,
		&u.IsAdmin,
		&u.Coins,
		&u.Reputation,
	)
	if err != nil {
		return User{}, err
	}

	return *u, nil
}

func (u *User) GetUserStatistics() (repStat float32, coinsStat float32, err error) {
	row := db.QueryRow(
		context.Background(),
		`SELECT
			(SELECT COUNT(*) FROM users WHERE reputation < $1) / (SELECT COUNT(*)::float FROM users) AS rep_stat,
			(SELECT COUNT(*) FROM users WHERE coins < $2) / (SELECT COUNT(*)::float FROM users) AS coins_stat`,
		u.Reputation,
		u.Coins,
	)
	err = row.Scan(
		&repStat,
		&coinsStat,
	)
	return
}

func (u *User) AddMoney(money int) {
	// for test only change it!
	_, _ = db.Exec(
		context.Background(),
		"UPDATE users SET coins = coins + $1 WHERE userid = $2",
		money,
		u.UserID,
	)
}

func (u *User) DecreaseMoney(money int) {
	// for test only change it!
	_, _ = db.Exec(
		context.Background(),
		"UPDATE users SET coins = coins - $1 WHERE userid = $2",
		money,
		u.UserID,
	)
}
