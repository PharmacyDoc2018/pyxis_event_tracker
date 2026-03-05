package cli

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
	"github.com/chzyer/readline"
	"github.com/joho/godotenv"
)

type Config struct {
	Db  *sql.DB
	Dbq *database.Queries
	Rl  *readline.Instance
}

func InitConfig() *Config {
	c := Config{}

	godotenv.Load(".env")
	connString := os.Getenv("CONNSTRING")

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		fmt.Printf("Error creating connection pool: %s\n ", err.Error())
	}
	c.Dbq = database.New(db)

	c.Rl = InitReadline()

	return &c
}
