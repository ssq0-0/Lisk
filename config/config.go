package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	RPC            map[string]string `json:"rpc"`
	ABIs           map[string]string `json:"abis"`
	TokenAddresses map[string]string `json:"token_addresses"`
	OkuAddresses   map[string]string `json:"oku_addresses"`
	IonicAddresses map[string]string `json:"ionic_addresses"`
	Endpoints      map[string]string `json:"enpoints"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
