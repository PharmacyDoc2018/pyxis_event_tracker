package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/internal/database"
	"github.com/joho/godotenv"
)

type config struct {
	db  *sql.DB
	dbq *database.Queries
}

func initConfig() *config {
	c := config{}

	godotenv.Load(".env")
	connString := os.Getenv("CONNSTRING")

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		fmt.Printf("Error creating connection pool: %s\n ", err.Error())
	}
	c.dbq = database.New(db)

	return &c
}
