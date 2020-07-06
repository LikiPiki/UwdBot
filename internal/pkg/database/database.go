package database

import (
	"context"
	"github.com/jackc/pgx/v4"
	"log"
	"os"
)

type Database struct {
	db            *pgx.Conn
	UserStorage   *UserStorage
	WeaponStorage *WeaponsStorage
}

func NewDatabase(context context.Context) (*Database, error) {
	db, err := pgx.Connect(context, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	log.Println("Success connected to database")

	return &Database{
		UserStorage:   NewUserStorage(db),
		WeaponStorage: NewWeaponsStorage(db),
	}, nil
}

func (d *Database) Close() error {
	return d.db.Close(context.Background())
}
