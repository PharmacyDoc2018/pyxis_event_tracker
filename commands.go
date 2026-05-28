package main

import (
	"fmt"
	"strconv"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
)

func (p *Process) setupCommands() {
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

		if p.dbConnection {
			p.findMissingPyxisEvents()
		}

		return nil

	}, cli.CommandArg{
		Name:     "name",
		Required: true,
	}, cli.CommandArg{
		Name:     "start_date",
		Required: true,
	})

	p.cliConfig.AddCommand("status", func(args []cli.CommandArg) error {
		p.logger.LogInfo("Status command executed")
		mode := p.state.Mode()
		fmt.Println(mode)
		p.logger.LogInfo(fmt.Sprintf("Current mode: %d", mode))
		return nil
	})

	p.cliConfig.AddCommand("add ERxItemId link", func(args []cli.CommandArg) error {
		p.logger.LogInfo("add ERxItemId link command executed")

		erx := ""
		itemId := ""

		for _, arg := range args {
			switch arg.Name {
			case "erx":
				erx = arg.Val

			case "itemid":
				itemId = arg.Val
			}
		}

		//-- Check for blank args
		if erx == "" {
			p.logger.LogError("Command failed: erx cannot be blank")
			return fmt.Errorf("error. erx cannot be blank")
		}
		if itemId == "" {
			p.logger.LogError("Command failed: itemid cannot be blank")
			return fmt.Errorf("error. itemid cannot be blank")
		}

		//-- Check for letters in args
		_, err := strconv.Atoi(erx)
		if err != nil {
			p.logger.LogError("Command failed: erx can only contain numbers")
			return fmt.Errorf("error. erx can only contain numbers")
		}
		_, err = strconv.Atoi(itemId)
		if err != nil {
			p.logger.LogError("Command failed: itemid can only contain numbers")
			return fmt.Errorf("error. itemid can only contain numbers")
		}

		//-- Call method to add link
		logErr := p.erxItemIdLinks.Add(erx, itemId)
		if logErr != nil {
			p.logger.LogError(logErr.LogError())
			return logErr
		}
		p.logger.LogInfo(fmt.Sprintf("Link created. ItemId %s now links to ERx %s", itemId, erx))
		fmt.Println("link created")
		return nil

	}, cli.CommandArg{
		Name:     "erx",
		Required: true,
	}, cli.CommandArg{
		Name:     "itemid",
		Required: true,
	})

}
