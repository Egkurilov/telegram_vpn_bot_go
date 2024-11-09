package main

import (
	"encoding/json"
	"io/ioutil"
)

type Messages struct {
	WelcomeMessage      string `json:"welcome_message"`
	OptionPrompt        string `json:"option_prompt"`
	ButtonOpenVPN       string `json:"button_OpenVPN"`
	ButtonOutline       string `json:"button_Outline"`
	ButtonTelegramProxy string `json:"button_TelegramProxy"`
	ButtonHttpProxy     string `json:"button_HttpProxy"`
	UnknownButton       string `json:"unknown_button"`
}

func loadMessages(filepath string) (Messages, error) {
	var messages Messages
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return messages, err
	}

	err = json.Unmarshal(data, &messages)
	if err != nil {
		return messages, err
	}

	return messages, nil
}
