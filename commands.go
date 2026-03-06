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
}
