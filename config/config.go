package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	RU        RU       `json:"ru"`
	NL        NL       `json:"nl"`
	Outline   Outline  `json:"outline"`
	OpenVPN   OpenVPN  `json:"openvpn"`
	HttpProxy []string `json:"users_proxy"`
}

type RU struct {
	Name   string `json:"name"`
	Server string `json:"server"`
	Secret string `json:"secret"`
}

type NL struct {
	Name   string `json:"name"`
	Server string `json:"server"`
}

type Outline struct {
	Server string `json:"server"`
	Token  string `json:"token"`
}

type OpenVPN struct {
	Script  string `json:"script"`
	Configs string `json:"configs"`
}

func LoadConfig() (*Config, error) {
	var config Config
	cfg, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(cfg, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
