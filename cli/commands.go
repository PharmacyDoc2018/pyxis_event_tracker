package cli

import (
	"fmt"
	"os"
)

func getCommands() map[string]cliCommand {
	commands := map[string]cliCommand{
		"hello": {
			name:        "hello",
			description: "greets the user",
			callback:    commandHello,
		},
		"exit": {
			name:        "exit",
			description: "exits the program",
			callback:    commandExit,
		},
		"add": {
			name:        "add",
			description: "adds element to tracker",
			callback:    commandAdd,
		},
	}

	return commands
}

func commandHello(c *Config) error {
	fmt.Println("Hello, World!")
	return nil
}

func commandExit(c *Config) error {
	fmt.Println("closing... goodbye!")
	c.Rl.Close()
	os.Exit(0)
	return nil
}

func commandAdd(c *Config) error {
	if len(c.lastInput) < 2 {
		return fmt.Errorf("error. too few arguments")
	}

	firstArg := c.lastInput[1]
	switch firstArg {
	case "pyxis":
		err := commandAddPyxis(c)
		if err != nil {
			return err
		}
	}
}
