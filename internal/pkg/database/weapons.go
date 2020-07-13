package database

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

type WeaponsStorage struct {
	*pgxpool.Pool
}

func NewWeaponsStorage(db *pgxpool.Pool) *WeaponsStorage {
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

func (w *WeaponsStorage) GetWeaponsCount(ctx context.Context) (int, error) {
	row := w.QueryRow(
		ctx,
		"SELECT COUNT (*) FROM  weapons",
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "cannot count all weapons")
	}

	return count, nil
}

func (w *WeaponsStorage) GetWeaponsLimitOffset(ctx context.Context, limit int, offset int) ([]Weapon, error) {
	rows, err := w.Query(
		ctx,
		"SELECT id, name, power, cost FROM weapons ORDER BY cost LIMIT $1 OFFSET $2",
		limit,
		offset,
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
