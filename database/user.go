package database

import (
	"context"
	"time"
)

type User struct {
	ID           uint64
	UserID       uint64
	Username     string
	Coins        int
	Reputation   int
	Blacklist    bool
	IsAdmin      bool
	WeaponsPower int
	ActiveDate   time.Time
	Activity     int
}

type Users []User

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
		"SELECT id, username, userid, blacklist, isadmin, coins, reputation, weapons_power, activ_date, activity FROM users WHERE userID = $1",
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
		&u.WeaponsPower,
		&u.ActiveDate,
		&u.Activity,
	)
	if err != nil {
		return User{}, err
	}

	return *u, nil
}

func (u *User) GetTopUsers(count int) (Users, error) {
	rows, err := db.Query(
		context.Background(),
		"SELECT id, username, userid, blacklist, isadmin, coins, reputation, weapons_power FROM users ORDER BY reputation desc, coins DESC LIMIT $1",
		count,
	)
	users := make(Users, 0)
	if err != nil {
		return Users{}, err
	}

	for rows.Next() {
		err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.UserID,
			&u.Blacklist,
			&u.IsAdmin,
			&u.Coins,
			&u.Reputation,
			&u.WeaponsPower,
		)

		if err != nil {
			return Users{}, err
		}
		users = append(users, *u)
	}
	return users, nil
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

func (u *User) AddPower(power int) {
	_, _ = db.Exec(
		context.Background(),
		"UPDATE users SET weapons_power = weapons_power + $1 WHERE id = $2",
		power,
		u.ID,
	)
}

func (u *User) AddMoneyToUsers(money int, us []int) error {
	_, err := db.Exec(
		context.Background(),
		"UPDATE users SET coins = coins + $1 WHERE userid = ANY($2)",
		money,
		us,
	)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) AddReputationToUsers(reputation int, us []int) error {
	_, err := db.Exec(
		context.Background(),
		"UPDATE users SET reputation = reputation + $1 WHERE userid = ANY($2)",
		reputation,
		us,
	)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) DecreaseMoneyToUsers(money int, us []int) error {
	_, err := db.Exec(
		context.Background(),
		"UPDATE users SET coins = coins - $1 WHERE userid = ANY($2)",
		money,
		us,
	)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) DecreaseReputationToUsers(reputation int, us []int32) error {
	_, err := db.Exec(
		context.Background(),
		"UPDATE users SET reputation = reputation - $1 WHERE userid = ANY($2)",
		reputation,
		us,
	)
	if err != nil {
		return err
	}
	return nil
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
