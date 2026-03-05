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
	lastInput []string
	commands  map[string]CliCommand
	Db        *sql.DB
	Dbq       *database.Queries
	Rl        *readline.Instance
}

func InitConfig() *Config {
	c := Config{}

	godotenv.Load(".env")
	connString := os.Getenv("CONNSTRING")

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		fmt.Printf("Error creating connection pool: %s\n ", err.Error())
	}
	c.Db = db

	c.Dbq = database.New(db)

	c.Rl = InitReadline()

	c.commands = getCommands()

	return &c
}

func (c *Config) commandLookup(input string) (CliCommand, error) {
	for _, command := range c.commands {
		if input == command.Name {
			return command, nil
		}
	}

	return CliCommand{}, fmt.Errorf("error. unknown command")
}

func (c *Config) CommandExe(input string) error {
	cleanInputAndStore(c, input)
	if len(c.lastInput) == 0 {
		return nil
	}

	command, err := c.commandLookup(c.lastInput[0])
	if err != nil {
		return err
	}

	err = command.Callback(c)
	if err != nil {
		return err
	}

	return nil

}
