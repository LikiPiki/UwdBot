package database

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type WeaponsStorage struct {
	*pgx.Conn
}

func NewWeaponsStorage(db *pgx.Conn) *WeaponsStorage {
	return &WeaponsStorage{db}
}

type Weapon struct {
	ID    int
	Name  string
	Power int
	Cost  int
}

func (w *WeaponsStorage) GetAllWeapons(ctx context.Context) ([]Weapon, error) {
	rows, err := w.Query(
		ctx,
		"SELECT id, name, power, cost FROM weapons ORDER BY cost",
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot select all weapons")
	}

	var weapons []Weapon
	for rows.Next() {
		var weapon Weapon
		err := rows.Scan(
			&weapon.ID,
			&weapon.Name,
			&weapon.Power,
			&weapon.Cost,
		)
		if err != nil {
			return nil, errors.Wrap(err, "cannot scan row")
		}

		weapons = append(weapons, weapon)
	}

	return weapons, nil
}

func (w *WeaponsStorage) GetWeaponsByID(ctx context.Context, id int) (Weapon, error) {
	row := w.QueryRow(
		ctx,
		"SELECT id, name, power, cost FROM weapons WHERE id = $1",
		id,
	)

	var weapon Weapon
	err := row.Scan(
		&weapon.ID,
		&weapon.Name,
		&weapon.Power,
		&weapon.Cost,
	)

	if err != nil {
		return Weapon{}, errors.Wrap(err, "cannot get weapon by id")
	}

	return weapon, nil
}
