package main

import (
	"context"
	"fmt"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
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
	p.setupCommands()

	//-- For GetPyxisEventsForDeviceByDateRange testing:
	params := database.GetPyxisEventsForDeviceByDateRangeParams{
		Device: "AUGUSTA2",
		Start:  "01/01/2026",
		End:    "02/01/2026",
	}

	events, err := p.dbq.GetPyxisEventsForDeviceByDateRange(context.Background(), params)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("printing %d events:\n", len(events))
		for _, event := range events {
			fmt.Printf("%s %s %s %s %s\n", event.TxDate.Time.Format("1/2/06"), event.TxTime.String, event.UserName.String, event.TransactionType.String, event.MedDisplayName.String)
		}
	}
	// --

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
