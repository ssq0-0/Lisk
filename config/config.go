package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	StartDate        string            `json:"start_date"`
	ActionCounts     int               `json:"actions_count"`
	MaxActionsTime   int               `json:"max_actions_time"`
	IonicBorrow      string            `json:"ionic_borrow_amount"`
	IonicSupply      string            `json:"ionic_supply_amount"`
	OkuPercentUsage  int               `json:"oku_percen_usage"`
	WrapAmount       string            `json:"amount_to_wrap"`
	AttentionGwei    string            `json:"attention_gwei"`
	AttentionTime    int               `json:"attention_time_cycle"`
	MaxAttentionTime int               `json:"max_attention_time"`
	StateFile        string            `json:"state_file"`
	RPC              map[string]string `json:"rpc"`
	ABIs             map[string]string `json:"abis"`
	TokenAddresses   map[string]string `json:"token_addresses"`
	OkuAddresses     map[string]string `json:"oku_addresses"`
	IonicAddresses   map[string]string `json:"ionic_addresses"`
	Endpoints        map[string]string `json:"enpoints"`
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
