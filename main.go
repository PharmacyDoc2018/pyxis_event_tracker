package main

import (
	"fmt"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	c := cli.InitConfig()
	defer c.Db.Close()

	err := c.Db.Ping()
	if err != nil {
		fmt.Println("warning: no database connection")
	}

	for {
		line, err := c.Rl.Readline()
		if err != nil {
			break
		}

		err = c.CommandExe(line)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println()
	}
}
