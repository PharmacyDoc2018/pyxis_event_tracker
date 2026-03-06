package main

import (
	"fmt"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	p := initProcess()
	defer p.db.Close()

	fmt.Println("Attempting to connect to database...")
	err := p.db.Ping()
	if err != nil {
		p.dbConnection = false
		fmt.Println("connection failed!")
		fmt.Println("warning: no database connection")
	} else {
		p.dbConnection = true
		fmt.Println("connection successful")
	}

	p.cliConfig = cli.InitConfig()

	for {
		line, err := p.cliConfig.Rl.Readline()
		if err != nil {
			break
		}

		err = p.cliConfig.CommandExe(line)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println()
	}
}
