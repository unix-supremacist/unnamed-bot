package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"bytes"
)

func genTTS(speaker, name, text string) []byte {
		type thing struct {
		Speaker string `json:"speaker"`
		Text string `json:"text"`
		Session string `json:"session"`
	}

	requestBody, err := json.Marshal(thing{
        Speaker: speaker,
		Text: text,
		Session: name,
    })
	res, err := http.Post(config.Ttsurl+"/tts/generate", "application/json", bytes.NewBuffer(requestBody))
	eror(err)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	eror(err)
	return body
}