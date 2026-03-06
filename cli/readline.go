package cli

import "github.com/chzyer/readline"

func InitReadline() *readline.Instance {

	completer := readline.NewPrefixCompleter(
		readline.PcItem("hi"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
	)

	rl, _ := readline.NewEx(&readline.Config{
		Prompt:       "> ",
		AutoComplete: completer,
	})

	return rl
}
