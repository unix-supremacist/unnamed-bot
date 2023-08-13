package main

import (
	"strings"
	"bytes"
	"math/rand"
	openai "github.com/sashabaranov/go-openai"
	discord "github.com/bwmarrin/discordgo"
	gfys "github.com/nyarumes/gofuckyourself"
)

var ds []*discord.Session

type DiscordConfig struct {
	Id, Userid, Token string
}

func newDiscordSession(bottoken string) *discord.Session {
	dg, err := discord.New("Bot " + bottoken)
	eror(err)
	dg.Identify.Intents = discord.IntentsGuildMessages | discord.IntentDirectMessages
	eror(dg.Open())
	ds = append(ds, dg)
	return dg
}

func newDiscordMod(dg *discord.Session,){
	newDiscordMessage := func(s *discord.Session, m *discord.MessageCreate){
		if m.Author.ID == s.State.User.ID {
			return
		}
		var filter *gfys.SwearFilter
		filter = gfys.NewSwearFilter(false, "test")
		caught, err := filter.Check(m.Content)
		eror(err)
		if len(caught) > 0 {
			logger.Println(m.Author.Username+": "+m.Content)
			logger.Println("caught: ", caught)
			s.ChannelMessageDelete(m.ChannelID, m.ID)
		}
	}
	dg.AddHandler(newDiscordMessage)
}

func newDiscordAi(botid, botname, botuser, speaker string, dg *discord.Session, responseChance, userBonus int, req openai.ChatCompletionRequest){
	newDiscordMessage := func(s *discord.Session, m *discord.MessageCreate){
		if m.Author.ID == s.State.User.ID {
			return
		}
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
			if channel, err = s.Channel(m.ChannelID); err != nil {
				eror(err)
			}
		}
		
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: m.Author.Username+": "+strings.Replace(m.Content, "<@"+botid+">", "@"+botname, -1),
		})
	
		user := responseChance
		if m.Author.ID == botuser {
			user += userBonus
		}

		mentioned := false
		for i := 0; i < len(m.Mentions); i++ {
			if m.Mentions[i].ID == s.State.User.ID {
				mentioned = true
			}
		}
	
		if mentioned || rand.Intn(100) <= user || channel.Type == discord.ChannelTypeDM {
			message := genMessage(m.Author.Username, strings.Replace(m.Content, "<@"+botid+">", "@"+botname, -1), botname, req)
			if message != "null" {
				if strings.Contains(m.Content, "voice") {
					s.ChannelFileSend(m.ChannelID, "voice_message.wav", bytes.NewReader(genTTS(speaker, botname, message)))
				} else {
					s.ChannelMessageSendReply(m.ChannelID, message, (*m).Reference())
				}
			}
		}
	}
	dg.AddHandler(newDiscordMessage)
}