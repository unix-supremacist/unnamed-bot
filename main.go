package main

import (
	"fmt"
	"log"
	"os"
	"encoding/json"
	"os/signal"
	"io/ioutil"
	"syscall"
)

var logger *log.Logger
var config Config

type Config struct {
	Url, Token string
	Bot []struct {
		Name string
		Discord DiscordConfig
		Twitch TwitchConfig
		Ai AiConfig
	}
	Channels []string
}

func main() {
	file, err := os.OpenFile("message.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	eror(err)
	defer file.Close()
	logger = log.New(file, "", 0)

	cs, err := ioutil.ReadFile("run.json")
	eror(err)
	eror(json.Unmarshal([]byte(cs), &config))

	filter()
	startAi()

	for i := 0; i < len(config.Bot); i++ {
		bot := config.Bot[i]
		
		ac := bot.Ai
		req := genReq(ac.Prompt)

		dc := bot.Discord
		
		if len(dc.Id) > 0 {
			dg := newDiscordSession(dc.Token)
			newDiscordMod(dg)
			if len(ac.Prompt) > 0 {
				newDiscordAi(dc.Id, bot.Name, dc.Userid, dg, ac.Chance, ac.Bonus, req)
			}
		}

		tc := bot.Twitch
		if len(tc.User) > 0 {
			tg := newTwitchSession(bot.Name, tc.Oauth)
			if len(ac.Prompt) > 0 {
				go newTwitchAi(bot.Name, tc.User, tc.Notme, tg, ac.Chance, ac.Bonus, req)
			}
			go func(){
				eror(tg.Connect())
			}()
		}
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	for i := 0; i < len(ds); i++ {
		ds[i].Close()
	}
}

func eror(err error) {
	if err != nil {
		fmt.Println(err)
	}
}