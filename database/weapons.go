package database

import "context"

type Weapon struct {
	ID    int
	Name  string
	Power int
	Cost  int
}

type Weapons []Weapon

func (w *Weapon) GetAllWeapons() (Weapons, error) {
	rows, err := db.Query(
		context.Background(),
		"SELECT id, name, power, cost FROM weapons ORDER BY cost",
	)
	weapons := make(Weapons, 0)
	weapon := Weapon{}
	if err != nil {
		return Weapons{}, nil
	}
	for rows.Next() {
		err := rows.Scan(
			&weapon.ID,
			&weapon.Name,
			&weapon.Power,
			&weapon.Cost,
		)
		if err != nil {
			return Weapons{}, err
		}
		weapons = append(weapons, weapon)
	}
	return weapons, nil
}

func (w *Weapon) GetWeaponsByID(id int) (Weapon, error) {
	row := db.QueryRow(
		context.Background(),
		"SELECT id, name, power, cost FROM weapons WHERE id = $1",
		id,
	)

	err := row.Scan(
		&w.ID,
		&w.Name,
		&w.Power,
		&w.Cost,
	)

	if err != nil {
		return *w, err
	}

	return *w, nil
}
