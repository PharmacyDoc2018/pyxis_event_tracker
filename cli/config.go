package cli

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
)

type Config struct {
	lastInput []string
	commands  map[string]cliCommand
	Rl        *readline.Instance
}

func InitConfig() *Config {
	c := Config{}
	c.Rl = InitReadline()
	c.commands = getCommands()

	return &c
}

type CommandArg struct {
	Name     string
	Val      any
	Required bool
}

type cliCommand struct {
	name     string
	function func() error
	args     []CommandArg
}

func (c *Config) AddCommand(name string, function func() error, args ...CommandArg) {
	newCommand := cliCommand{
		name:     name,
		function: function,
		args:     args,
	}

	c.commands[name] = newCommand
}

func (c *Config) commandLookup(input string) (cliCommand, bool) {
	cmd, ok := c.commands[input]
	return cmd, ok
}

func (c *Config) CommandExe(input string) error {
	cleanInputAndStore(c, input)

	if len(c.lastInput) == 0 {
		return nil
	}

	cmdKey := c.parseInputForCommand()
	if cmdKey == "" {
		return fmt.Errorf("error. no commands found")
	}

	cmd, ok := c.commandLookup(cmdKey)
	if !ok {
		return fmt.Errorf("error. command %s not found", cmdKey)
	}

	//-- add arg parsing function

	return nil

}

func (c *Config) parseInputForCommand() string {
	nonArgs := []string{}

	for _, i := range c.lastInput {
		if !strings.Contains(i, "=") {
			nonArgs = append(nonArgs, i)
		}
	}

	if len(nonArgs) == 0 {
		return ""
	}

	return strings.Join(nonArgs, " ")

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
