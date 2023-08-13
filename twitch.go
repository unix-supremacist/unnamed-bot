package main

import (
	"math/rand"
	"strings"
	"github.com/gempir/go-twitch-irc/v4"
	openai "github.com/sashabaranov/go-openai"
)

type TwitchConfig struct {
	Oauth, User, Notme string
}

func newTwitchSession(username, oauth string) *twitch.Client {
	tc := twitch.NewClient(username, "oauth:"+oauth)
	for i := 0; i < len(config.Channels); i++ {
		tc.Join(config.Channels[i])
	}
	return tc
}

func newTwitchAi(username, usern, notme string, tc *twitch.Client, chance, bonus int, req openai.ChatCompletionRequest){
	tc.OnPrivateMessage(func(message twitch.PrivateMessage) {
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: message.User.Name+": "+message.Message,
		})

		user := chance
		if strings.ToLower(message.User.Name) == usern {
			user += bonus
		}
		if strings.ToLower(message.User.Name) == notme && rand.Intn(100) >= user {
			return
		}

		if strings.Contains(strings.ToLower(message.Message), "@"+username) || rand.Intn(100) <= user == true {
			gmessage := genMessage(message.User.Name, message.Message, username, req)
			if gmessage != "null" {
				tc.Say(message.Channel, "@"+message.User.Name+" "+gmessage)
			}
		}
	})
}