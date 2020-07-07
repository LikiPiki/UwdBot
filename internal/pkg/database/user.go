package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

const (
	PerDayActivity = 5
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

func (u *UserStorage) GetUserStatistics(ctx context.Context, rep int, cn int) (reputation int, coins int, err error) {
	row := u.QueryRow(
		ctx,
		`SELECT
			CAST((SELECT COUNT(*) FROM users WHERE reputation < $1) / (SELECT COUNT(*)::float FROM users) AS int) AS rep_stat,
			CAST((SELECT COUNT(*) FROM users WHERE coins < $2) / (SELECT COUNT(*)::float FROM users) AS int) AS coins_stat`,
		rep,
		cn,
	)

	var repStat int
	var coinsStat int

	if err := row.Scan(
		&repStat,
		&coinsStat,
	); err != nil {
		return 0, 0, errors.Wrap(err, "cannot get user statistics")
	}

	return repStat, coinsStat, nil
}

func (u *UserStorage) AddMoney(ctx context.Context, userID uint64, money int) error {
	// for test only change it!
	_, err := u.Exec(
		ctx,
		"UPDATE users SET coins = coins + $1 WHERE userid = $2",
		money,
		userID,
	)
	if err != nil {
		return errors.Wrap(err, "cannot add money")
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
	_, err := u.Exec(
		ctx,
		"UPDATE users SET weapons_power = weapons_power + $1 WHERE userid = $2",
		power,
		userID,
	)
	if err != nil {
		return errors.Wrap(err, "cannot add power")
	}

	return nil
}

func (u *UserStorage) AddMoneyToUsers(ctx context.Context, money int, us []int) error {
	_, err := u.Exec(
		ctx,
		"UPDATE users SET coins = coins + $1 WHERE userid = ANY($2)",
		money,
		us,
	)
	if err != nil {
		return errors.Wrap(err, "cannot add money to users")
	}

	return nil
}

func (u *UserStorage) AddReputationToUsers(ctx context.Context, reputation int, us []int) error {
	_, err := u.Exec(
		ctx,
		"UPDATE users SET reputation = reputation + $1 WHERE userid = ANY($2)",
		reputation,
		us,
	)
	if err != nil {
		return errors.Wrap(err, "cannot add reputation to users")
	}

	return nil
}

func (u *UserStorage) DecreaseMoneyToUsers(ctx context.Context, money int, us []int) error {
	_, err := u.Exec(
		ctx,
		"UPDATE users SET coins = coins - $1 WHERE userid = ANY($2)",
		money,
		us,
	)
	if err != nil {
		return errors.Wrap(err, "cannot decrease money from users")
	}

	return nil
}

func (u *UserStorage) DecreaseReputationToUsers(ctx context.Context, reputation int, us []int) error {
	_, err := u.Exec(
		ctx,
		"UPDATE users SET reputation = reputation - $1 WHERE userid = ANY($2)",
		reputation,
		us,
	)
	if err != nil {
		return errors.Wrap(err, "cannot decrease rep for users")
	}

	return nil
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