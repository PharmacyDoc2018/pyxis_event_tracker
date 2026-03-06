package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
	"github.com/joho/godotenv"
)

type ProcessState struct {
	pyxisUnits   []string
	db           *sql.DB
	dbq          *database.Queries
	cliConfig    *cli.Config
	dbConnection bool
}

func initProcess() *ProcessState {
	p := ProcessState{}

	godotenv.Load(".env")
	connString := os.Getenv("CONNSTRING")

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		fmt.Printf("Error creating connection pool: %s\n ", err.Error())
	}
	p.db = db

	p.dbq = database.New(db)

	return &p
}
