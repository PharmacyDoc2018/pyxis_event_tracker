package main

import (
	"fmt"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	c := cli.InitConfig()
	defer c.Db.Close()

	for {
		line, err := c.Rl.Readline()
		if err != nil {
			break
		}

		fmt.Println(line)
		fmt.Println()
	}
}
