package main

import (
	"fmt"

	"github.com/wyt-labs/wyt-core/internal/pkg/config"
	"github.com/wyt-labs/wyt-core/pkg/extension"
)

type Chat struct {
	chatgptDriver *extension.ChatgptDriver
	HistoryMsgs   []*extension.ChatgptMsg
}

func NewChat() (*Chat, error) {
	chatgptDriver, err := extension.NewChatgptDriver(&extension.ChatgptConfig{
		Endpoint:        "https://sme.openai.azure.com/",
		APIKey:          "b19a2b2fdae04ff2ae0f6cdc4792eb86",
		Model:           "gpt4-turbo",
		Temperature:     0.7,
		PresencePenalty: 0,
	}, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Chat{
		chatgptDriver: chatgptDriver,
	}, nil
}

func (c *Chat) Executor(msg string) {
	res, err := func() (string, error) {
		c.HistoryMsgs = append(c.HistoryMsgs, &extension.ChatgptMsg{
			Role:    "user",
			Content: msg,
		})

		allMsgs := []*extension.ChatgptMsg{{
			Role:    "system",
			Content: config.ChatPrompt,
		}}
		historyMsgsSize := len(c.HistoryMsgs)
		if historyMsgsSize <= 5 {
			allMsgs = append(allMsgs, c.HistoryMsgs...)
		} else if historyMsgsSize > 5 {
			allMsgs = append(allMsgs, c.HistoryMsgs[historyMsgsSize-5:historyMsgsSize]...)
		}

		return c.chatgptDriver.ChatCompletions(allMsgs)
	}()
	if err != nil {
		fmt.Println("Something went wrong:", err.Error())
	} else {
		fmt.Println(res)
	}
}
