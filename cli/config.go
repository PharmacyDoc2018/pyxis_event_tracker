package cli

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
)

type Config struct {
	lastInput      []string
	commands       map[string]cliCommand
	rlAutoComplete *readline.PrefixCompleter
	Rl             *readline.Instance
}

func InitConfig() *Config {
	c := Config{}
	c.Rl, c.rlAutoComplete = InitReadline()
	c.commands = map[string]cliCommand{}

	return &c
}

type CommandArg struct {
	Name string
	Val  any
}

type cliCommand struct {
	name     string
	function func([]CommandArg) error
	args     []CommandArg
}

func (c *Config) AddCommand(name string, function func([]CommandArg) error, args ...CommandArg) {
	newCommand := cliCommand{
		name:     name,
		function: function,
		args:     args,
	}

	c.commands[name] = newCommand

	c.rlAutoComplete.Children = append(c.rlAutoComplete.Children,
		readline.PcItem(name),
	)

	argsChildren := []readline.PrefixCompleterInterface{}
	for _, arg := range args {
		argsChildren = append(argsChildren,
			readline.PcItem(arg.Name+"="),
		)
	}
	c.rlAutoComplete.Children[len(c.rlAutoComplete.Children)-1].SetChildren(argsChildren)
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
	args := c.parseInputForArgs()

	err := cmd.function(args)
	if err != nil {
		return err
	}

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

func (c *Config) parseInputForArgs() []CommandArg {
	args := []string{}

	for _, i := range c.lastInput {
		if strings.Contains(i, "=") {
			args = append(args, i)
		}
	}

	commandArgs := []CommandArg{}
	for _, arg := range args {
		splitArg := strings.Split(arg, "=")
		commandArgs = append(commandArgs, CommandArg{
			Name: splitArg[0],
			Val:  splitArg[1],
		})
	}

	return commandArgs
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
