package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"encoding/json"
	"os/signal"
	"io/ioutil"
	"math/rand"
	"syscall"
	"strings"
	"github.com/gempir/go-twitch-irc/v4"
	openai "github.com/sashabaranov/go-openai"
	discord "github.com/bwmarrin/discordgo"
	gfys "github.com/nyarumes/gofuckyourself"
)

var aic *openai.Client
var logger *log.Logger
var config Config
var aifilter *gfys.SwearFilter
var ds []*discord.Session

type Config struct {
	Url, Token string
	Ai []struct {
		Prompt, Name string
		Discord DiscordConfig
		Twitch TwitchConfig
	}
	Channels []string
}

type DiscordConfig struct {
	Id, Userid, Token string
	Chance, Bonus int
}

type TwitchConfig struct {
	Oauth, User, Notme string
}

type Filter struct {
	Words []string
}

func main() {
	file, err := os.OpenFile("message.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	eror(err)
	defer file.Close()
	logger = log.New(file, "", 0)

	cs, err := ioutil.ReadFile("run.json")
	eror(err)
	eror(json.Unmarshal([]byte(cs), &config))

	af, err := ioutil.ReadFile("aifilter.json")
	eror(err)
	var aff Filter
	eror(json.Unmarshal([]byte(af), &aff))
	aifilter = gfys.NewSwearFilter(false, aff.Words...)

	aiconfig := openai.DefaultConfig(config.Token)
	aiconfig.BaseURL=config.Url
	aic = openai.NewClientWithConfig(aiconfig)

	for i := 0; i < len(config.Ai); i++ {
		ai := config.Ai[i]
		req := genReq(ai.Prompt)

		dc := ai.Discord
		if len(dc.Id) > 0 {
			go newDiscordAi(dc.Id, ai.Name, dc.Userid, dc.Token, dc.Chance, dc.Bonus, req)
		}

		tc := ai.Twitch
		if len(tc.User) > 0 {
			go newTwitchAi(ai.Name, tc.Oauth, tc.User, tc.Notme, req)
		}
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	for i := 0; i < len(ds); i++ {
		ds[i].Close()
	}
}

func newTwitchAi(username, oauth, usern, notme string, req openai.ChatCompletionRequest){
	tc := twitch.NewClient(username, "oauth:"+oauth)
	for i := 0; i < len(config.Channels); i++ {
		tc.Join(config.Channels[i])
	}
	tc.OnPrivateMessage(func(message twitch.PrivateMessage) {
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: message.User.Name+": "+message.Message,
		})

		user := 7
		if strings.ToLower(message.User.Name) == usern {
			user += 13
		}
		if strings.ToLower(message.User.Name) == notme && rand.Intn(100) >= user {
			return
		}

		if strings.Contains(strings.ToLower(message.Message), "@"+username) || rand.Intn(100) <= user == true {
			tc.Say(message.Channel, "@"+message.User.Name+" "+genMessage(message.User.Name, message.Message, username, req))
		}
	})
	eror(tc.Connect())
}

func newDiscordAi(botid, botname, botuser, bottoken string, responseChance, userBonus int, req openai.ChatCompletionRequest){
	newDiscordMessage := func(s *discord.Session, m *discord.MessageCreate){
		if m.Author.ID == s.State.User.ID {
			return
		}
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: m.Author.Username+": "+strings.Replace(m.Content, "<@"+botid+">", "@"+botname, -1),
		})
	
		user := responseChance
		if m.Author.ID == botuser {
			user += userBonus
		}
	
		if strings.Contains(m.Content, "<@"+botid+">") || rand.Intn(100) <= user == true {
			s.ChannelMessageSend(m.ChannelID, "<@!"+m.Author.ID+"> "+genMessage(m.Author.Username, strings.Replace(m.Content, "<@"+botid+">", "@"+botname, -1), botname, req))
		}
	}

	dg, err := discord.New("Bot " + bottoken)
	eror(err)
	dg.AddHandler(newDiscordMessage)
	dg.Identify.Intents = discord.IntentsGuildMessages | discord.IntentDirectMessages
	eror(dg.Open())
	ds = append(ds, dg)
}

func eror(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func genMessage(username, message, botname string, req openai.ChatCompletionRequest) string {
	logger.Println(username+": "+message)
	resp, err := aic.CreateChatCompletion(context.Background(), req)
	eror(err)
	if resp.Choices == nil {
		return "null"
	}
	resp.Choices[0].Message.Content = strings.Replace(resp.Choices[0].Message.Content, "</s>", "", -1)
	logger.Println(botname+": "+resp.Choices[0].Message.Content)
	caught, err := aifilter.Check(resp.Choices[0].Message.Content)
	eror(err)
	if len(caught) > 0 {
		logger.Println("caught: ", caught)
		resp.Choices[0].Message.Content = "filtered."
	}
	req.Messages = append(req.Messages, resp.Choices[0].Message)
	return resp.Choices[0].Message.Content
}

func genReq(prompt string) openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
		},
	}
}