package main

import (
	"context"
	"strings"
	"encoding/json"
	"io/ioutil"
	openai "github.com/sashabaranov/go-openai"
	gfys "github.com/nyarumes/gofuckyourself"
)

var aic *openai.Client
var aifilter *gfys.SwearFilter

type AiConfig struct {
	Prompt, Speaker string
	Chance, Bonus int
}

type Filter struct {
	Words []string
}

func filter(){
	af, err := ioutil.ReadFile("aifilter.json")
	eror(err)
	var aff Filter
	eror(json.Unmarshal([]byte(af), &aff))
	aifilter = gfys.NewSwearFilter(false, aff.Words...)
}

func startAi(){
	aiconfig := openai.DefaultConfig(config.Token)
	aiconfig.BaseURL=config.Url
	aic = openai.NewClientWithConfig(aiconfig)
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