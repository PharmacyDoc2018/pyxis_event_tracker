package main

import (
	"fmt"
	"os"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
)

func setupCommands(c *cli.Config) {
	c.AddCommand("hello", func([]cli.CommandArg) error {
		fmt.Println("Hello, World!")
		return nil
	})

	c.AddCommand("exit", func([]cli.CommandArg) error {
		fmt.Println("closing... goodbye!")
		c.Rl.Close()
		os.Exit(0)
		return nil
	})

	c.AddCommand("echo", func(args []cli.CommandArg) error {
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

}
