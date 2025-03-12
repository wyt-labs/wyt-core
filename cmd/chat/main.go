package main

import (
	"fmt"

	"github.com/c-bata/go-prompt"
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "Show me some info about ", Description: ""},
		{Text: "Compare projects between ", Description: ""},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func main() {
	fmt.Print("Welcome to wyt-chat, I'm ChatGPt-based ai assistant, you can ask me questions about web3 projects.\n")
	fmt.Println("Please use `exit` or `Ctrl-D` to exit this program.")

	c, err := NewChat()
	if err != nil {
		panic(err)
	}
	p := prompt.New(
		c.Executor,
		completer,
		prompt.OptionTitle("wyt-chat"),
		prompt.OptionPrefix(">>> "),
		prompt.OptionInputTextColor(prompt.Yellow),
	)
	p.Run()
}
