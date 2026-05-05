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
		switch len(args) {
		case 0:
			return fmt.Errorf("error. arguments missing. enter \"add pyxis help\" for format")

		case 1:
			switch args[0].Val {
			case "help":
				fmt.Println("The \"add pyxis\" command adds a Pyxis unit to the list of units to track.")
				fmt.Println("Example command format:")
				fmt.Println("add pyxis name=[name] start=[start date]")
				return nil

			default:
				return fmt.Errorf("error. arguments missing. enter \"add pyxis help\" for format")

			}

		case 3:
			return fmt.Errorf("error. too many arguments. enter \"add pyxis help\" for format")

		}

		return nil

	})

}
