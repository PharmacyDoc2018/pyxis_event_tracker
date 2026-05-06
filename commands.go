package main

import (
	"fmt"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
)

func (p *ProcessState) setupCommands() {
	p.cliConfig.AddCommand("hello", func([]cli.CommandArg) error {
		fmt.Println("Hello, World!")
		p.logger.LogInfo("Command hello executed")
		return nil
	})

	p.cliConfig.AddCommand("exit", func([]cli.CommandArg) error {
		p.logger.LogInfo("Command exit executed")
		fmt.Println("closing... goodbye!")
		p.exit()
		return nil
	})

	p.cliConfig.AddCommand("echo", func(args []cli.CommandArg) error {
		if len(args) == 0 {
			return fmt.Errorf("error. need phrase argument")
		}

		if len(args) > 1 {
			return fmt.Errorf("error. too many arguments")
		}

		fmt.Println(args[0].Val)
		return nil
	}, cli.CommandArg{
		Name: "phrase",
	})

	p.cliConfig.AddCommand("add pyxis", func(args []cli.CommandArg) error {
		p.logger.LogInfo("Add pyxis command executed")
		pyxisName := ""
		startDateString := ""

		for _, arg := range args {
			switch arg.Name {
			case "name":
				pyxisName = arg.Val

			case "start_date":
				startDateString = arg.Val
			}
		}

		startDate, err := parseDate(startDateString)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error. Unable to parse start date for new Pyxis event log: %s", err.Error()))
			return err
		}

		err = p.createNewPyxisEventLog(pyxisName, startDate)
		if err != nil {
			return err
		}

		return nil

	}, cli.CommandArg{
		Name:     "name",
		Required: true,
	}, cli.CommandArg{
		Name:     "start_date",
		Required: true,
	})

}
