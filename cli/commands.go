package cli

import (
	"fmt"
	"os"
	"strings"
)

type CliCommand struct {
	Name        string
	Description string
	Callback    func(*Config) error
}

func getCommands() map[string]CliCommand {
	commands := map[string]CliCommand{
		"hello": {
			Name:        "hello",
			Description: "greets the user",
			Callback:    commandHello,
		},
		"exit": {
			Name:        "exit",
			Description: "exits the program",
			Callback:    commandExit,
		},
	}

	return commands
}

func cleanInput(text string) []string {
	var textWords []string
	text = strings.TrimSpace(text)
	firstPass := strings.Split(text, " ")

	for _, word := range firstPass {
		if word != "" {
			textWords = append(textWords, word)
		}
	}

	return textWords
}

func cleanInputAndStore(c *Config, input string) {
	c.lastInput = cleanInput(input)
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
