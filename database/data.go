package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

var db *pgx.Conn

// InitDB - create database connection
func InitDB() {
	var err error
	db, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Success connected to database")
	}
}

func CloseDatabase() {
	db.Close(context.Background())
}

// InsertError - and error inserting to database
type InsertError struct{}

func (e *InsertError) Error() string {
	return fmt.Sprintf("Can't insert to database")
}
