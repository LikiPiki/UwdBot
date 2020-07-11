package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

const (
	PerDayActivity = 15
)

type UserStorage struct {
	*pgx.Conn
}

func NewUserStorage(db *pgx.Conn) *UserStorage {
	return &UserStorage{db}
}

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

func needDateUpdate(old time.Time, new time.Time) bool {
	return !(old.Day() == new.Day() && (old.Month() == new.Month()))
}

func (u *UserStorage) SwitchBanUser(ctx context.Context, username string, banState bool) error {
	commandTag, err := u.Exec(
		ctx,
		"UPDATE users SET blacklist = $1 where username = $2",
		banState,
		username,
	)

	if err != nil {
		return errors.Wrap(err, "cannot switch ban user")
	}

	if commandTag.RowsAffected() != 1 {
		return errors.New("cannot switch ban user")
	}
	return nil
}

func (u *UserStorage) CreateNewUser(ctx context.Context, username string, userID uint64) (uint64, error) {
	row := u.QueryRow(
		ctx,
		"INSERT INTO users (username, userID, coins) VALUES ($1, $2, 100) RETURNING id",
		username,
		userID,
	)

	var ID uint64
	err := row.Scan(&ID)
	if err != nil {
		return 0, errors.Wrap(err, "cannot create new user")
	}

	return ID, nil
}

func (u *UserStorage) CountUsersWithID(ctx context.Context, id int) (int, error) {
	row := u.QueryRow(
		ctx,
		"SELECT COUNT (*) FROM users WHERE userID = $1",
		id,
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "cannot count users")
	}

	return count, nil
}

func (u *UserStorage) DeleteUser(ctx context.Context, id int) (int, error) {
	commandTag, err := u.Exec(
		ctx,
		"DELETE FROM users WHERE userID = $1",
		id,
	)

	if err != nil {
		return 0, errors.Wrap(err, "cannot delete user")
	}

	if commandTag.RowsAffected() != 1 {
		return 0, errors.New("no row found to delete")
	}

	return int(commandTag.RowsAffected()), nil
}

func (u *UserStorage) UpdateActivity(ctx context.Context, user *User) (int, error) {
	if needDateUpdate(user.ActiveDate, time.Now()) {
		commandTag, err := u.Exec(
			ctx,
			"UPDATE users SET activ_date = CURRENT_TIMESTAMP, activity = $2 where userid = $1",
			user.UserID,
			PerDayActivity,
		)
		if err != nil {
			return 0, errors.Wrap(err, "cannot update user activity")
		}

		if commandTag.RowsAffected() != 1 {
			return 0, errors.New("no row found to update")
		}
		user.Activity = PerDayActivity
	}
	return user.Activity, nil
}

func (u *UserStorage) DecreaseActivity(ctx context.Context, userID int) error {
	user, err := u.FindUserByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "cannot find user")
	}

	if user.Activity > 0 {
		commandTag, err := u.Exec(
			ctx,
			"UPDATE users SET activity = $1 WHERE userID = $2",
			user.Activity-1,
			user.UserID,
		)
		if err != nil {
			return errors.Wrap(err, "cannot set new activity")
		}

		if commandTag.RowsAffected() != 1 {
			return errors.New("no row found to update")
		}

		return nil
	}
	return nil
}

func (u *UserStorage) FindUserByID(ctx context.Context, id int) (User, error) {
	row := u.QueryRow(
		ctx,
		"SELECT id, username, userid, blacklist, isadmin, coins, reputation, weapons_power, activ_date, activity FROM users WHERE userID = $1",
		id,
	)

	var user User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.UserID,
		&user.Blacklist,
		&user.IsAdmin,
		&user.Coins,
		&user.Reputation,
		&user.WeaponsPower,
		&user.ActiveDate,
		&user.Activity,
	)
	if err != nil {
		return User{}, errors.Wrap(err, "cannot find user")
	}

	user.Activity, err = u.UpdateActivity(ctx, &user)
	if err != nil {
		return User{}, errors.Wrap(err, "cannot update user activity")
	}

	return user, nil
}

func (u *UserStorage) FindUserByUsername(ctx context.Context, username string) (User, error) {
	row := u.QueryRow(
		ctx,
		"SELECT id, username, userid, blacklist, isadmin, coins, reputation, weapons_power, activ_date, activity FROM users WHERE username = $1",
		username,
	)

	var user User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.UserID,
		&user.Blacklist,
		&user.IsAdmin,
		&user.Coins,
		&user.Reputation,
		&user.WeaponsPower,
		&user.ActiveDate,
		&user.Activity,
	)
	if err != nil {
		return User{}, errors.Wrap(err, "cannot find user by name")
	}

	user.Activity, err = u.UpdateActivity(ctx, &user)
	if err != nil {
		return User{}, errors.Wrap(err, "cannot update user activity")
	}

	return user, nil
}

func (u *UserStorage) GetTopUsers(ctx context.Context, count int) ([]User, error) {
	rows, err := u.Query(
		ctx,
		"SELECT id, username, userid, blacklist, isadmin, coins, reputation, weapons_power FROM users ORDER BY reputation desc, coins DESC LIMIT $1",
		count,
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get top users")
	}

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.UserID,
			&user.Blacklist,
			&user.IsAdmin,
			&user.Coins,
			&user.Reputation,
			&user.WeaponsPower,
		)

		if err != nil {
			return nil, errors.Wrap(err, "cannot scan top users row")
		}

		users = append(users, user)
	}
	return users, nil
}

func (u *UserStorage) GetUserStatistics(ctx context.Context, rep int, cn int) (reputation float32, coins float32, err error) {
	row := u.QueryRow(
		ctx,
		`SELECT
			CAST((SELECT COUNT(*) FROM users WHERE reputation <= $1) / (SELECT COUNT(*)::float FROM users) AS float) AS rep_stat,
			CAST((SELECT COUNT(*) FROM users WHERE coins <= $2) / (SELECT COUNT(*)::float FROM users) AS float) AS coins_stat`,
		rep,
		cn,
	)

	var repStat, coinsStat float32

	if err := row.Scan(
		&repStat,
		&coinsStat,
	); err != nil {
		return 0, 0, errors.Wrap(err, "cannot get user statistics")
	}

	return repStat, coinsStat, nil
}

func (u *UserStorage) AddMoney(ctx context.Context, userID uint64, money int) error {
	commandTag, err := u.Exec(
		ctx,
		"UPDATE users SET coins = coins + $1 WHERE userid = $2",
		money,
		userID,
	)
	if err != nil {
		return errors.Wrap(err, "cannot add money")
	}

	if commandTag.RowsAffected() != 1 {
		return errors.New("cannot add money to user")
	}

	return nil
}

func (u *UserStorage) AddReputation(ctx context.Context, userID uint64, reputation int) error {
	commandTag, err := u.Exec(
		ctx,
		"UPDATE users SET reputation = reputation + $1 WHERE userid = $2",
		reputation,
		userID,
	)

	if err != nil {
		return errors.Wrap(err, "cannot add money")
	}

	if commandTag.RowsAffected() != 1 {
		return errors.New("cannot add reputation")
	}

	return nil
}

func (u *UserStorage) AddPower(ctx context.Context, userID int, power int) error {
	commandTag, err := u.Exec(
		ctx,
		"UPDATE users SET weapons_power = weapons_power + $1 WHERE userid = $2",
		power,
		userID,
	)
	if err != nil {
		return errors.Wrap(err, "cannot add power")
	}

	if commandTag.RowsAffected() != 1 {
		return errors.New("cannot add power to user")
	}

	return nil
}

func (u *UserStorage) AddMoneyToUsers(ctx context.Context, money int, us []int) error {
	commandTag, err := u.Exec(
		ctx,
		"UPDATE users SET coins = coins + $1 WHERE userid = ANY($2)",
		money,
		us,
	)
	if err != nil {
		return errors.Wrap(err, "cannot add money to users")
	}

	if commandTag.RowsAffected() != int64(len(us)) {
		return errors.New(
			fmt.Sprintf(
				"cannot add money to users row affected %d, expected %d",
				commandTag.RowsAffected(),
				len(us),
			),
		)
	}

	return nil
}

func (u *UserStorage) AddReputationToUsers(ctx context.Context, reputation int, us []int) error {
	commandTag, err := u.Exec(
		ctx,
		"UPDATE users SET reputation = reputation + $1 WHERE userid = ANY($2)",
		reputation,
		us,
	)
	if err != nil {
		return errors.Wrap(err, "cannot add reputation to users")
	}

	if commandTag.RowsAffected() != int64(len(us)) {
		return errors.New("cannot add reputation to users")
	}

	return nil
}

func (u *UserStorage) DecreaseMoneyToUsers(ctx context.Context, money int, us []int) error {
	commandTag, err := u.Exec(
		ctx,
		"UPDATE users SET coins = coins - $1 WHERE userid = ANY($2)",
		money,
		us,
	)
	if err != nil {
		return errors.Wrap(err, "cannot decrease money from users")
	}

	if commandTag.RowsAffected() != int64(len(us)) {
		return errors.New("cannot add reputation to users")
	}

	return nil
}

func (u *UserStorage) DecreaseReputationToUsers(ctx context.Context, reputation int, us []int) error {
	commandTag, err := u.Exec(
		ctx,
		"UPDATE users SET reputation = reputation - $1 WHERE userid = ANY($2)",
		reputation,
		us,
	)
	if err != nil {
		return errors.Wrap(err, "cannot decrease rep for users")
	}

	if commandTag.RowsAffected() != int64(len(us)) {
		return errors.New("cannot add reputation to users")
	}

	return nil
}

func (u *UserStorage) FindUsersPowerBetween(ctx context.Context, userID uint64, min int, max int, limit int) ([]User, error) {
	rows, err := u.Query(
		ctx,
		"SELECT userid, username, weapons_power FROM users WHERE (weapons_power BETWEEN $1 AND $2) AND userid != $3 LIMIT $4",
		min,
		max,
		userID,
		limit,
	)
	if err != nil {
		return []User{}, errors.Wrap(err, "cannot select users between")
	}

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.WeaponsPower,
		)

		if err != nil {
			return nil, errors.Wrap(err, "cannot scan between users row")
		}

		users = append(users, user)
	}

	return users, nil
}

func (u *UserStorage) DecreaseMoney(ctx context.Context, userID uint64, money int) error {
	commandTag, err := u.Exec(
		ctx,
		"UPDATE users SET coins = coins - $1 WHERE userid = $2",
		money,
		userID,
	)

	if err != nil {
		return errors.Wrap(err, "cannot decrease money")
	}

	if commandTag.RowsAffected() != 1 {
		return errors.New("no row found to update")
	}

	return nil
}

func (u *UserStorage) IsAdmin(ctx context.Context, ID int) bool {
	user, err := u.FindUserByID(ctx, ID)
	if err != nil {
		return false
	}

	return user.IsAdmin
}
