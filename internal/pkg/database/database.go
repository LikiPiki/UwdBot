package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Database struct {
	db            *pgxpool.Pool
	UserStorage   *UserStorage
	WeaponStorage *WeaponsStorage
	GifsStorage   *GifsStorage
}

func NewDatabase(context context.Context) (*Database, error) {
	db, err := pgxpool.Connect(context, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	log.Println("Success connected to database")

	return &Database{
		UserStorage:   NewUserStorage(db),
		WeaponStorage: NewWeaponsStorage(db),
		GifsStorage:   NewGifsStorage(db),
	}, nil
}

func (d *Database) Close() {
	d.db.Close()
}
