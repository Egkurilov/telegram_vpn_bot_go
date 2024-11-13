package main

import (
	"encoding/json"
	"os"
)

type Messages struct {
	WelcomeMessage      string `json:"welcome_message"`
	OptionPrompt        string `json:"option_prompt"`
	ButtonOpenVPN       string `json:"button_OpenVPN"`
	ButtonopenvpnNl     string `json:"button_OpenVPN_NL"`
	ButtonopenvpnRu     string `json:"button_OpenVPN_RU"`
	ButtonOutline       string `json:"button_Outline"`
	ButtonTelegramProxy string `json:"button_TelegramProxy"`
	ButtonHttpProxy     string `json:"button_HttpProxy"`
	UnknownButton       string `json:"unknown_button"`
}

func loadMessages(filepath string) (Messages, error) {
	var messages Messages
	data, err := os.ReadFile(filepath)
	if err != nil {
		return messages, err
	}

	err = json.Unmarshal(data, &messages)
	if err != nil {
		return messages, err
	}

	return messages, nil
}
