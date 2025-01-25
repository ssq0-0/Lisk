package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Threads           int               `json:"threads"`
	StartDate         string            `json:"start_date"`
	ActionCounts      int               `json:"actions_count"`
	MaxActionsTime    int               `json:"max_actions_time"`
	IonicBorrow       string            `json:"ionic_borrow_amount"`
	IonicSupply       string            `json:"ionic_supply_amount"`
	OkuPercentUsage   int               `json:"oku_percen_usage"`
	WrapMinAmount     string            `json:"min_amount_to_wrap"`
	WrapMaxAmount     string            `json:"max_amount_to_wrap"`
	SwapUSDTMinAmount string            `json:"min_usdc_amount_to_swap"`
	SwapUSDTMaxAmount string            `json:"max_usdc_amount_to_swap"`
	SwapEthMaxAmount  string            `json:"max_eth_amount_to_swap"`
	SwapEthMinAmount  string            `json:"min_eth_amount_to_swap"`
	MinUSDTForSwap    string            `json:"min_usdt_amount_to_swap"`
	AttentionGwei     string            `json:"attention_gwei"`
	AttentionTime     int               `json:"attention_time_cycle"`
	MaxAttentionTime  int               `json:"max_attention_time"`
	StateFile         string            `json:"state_file"`
	RPC               map[string]string `json:"rpc"`
	ABIs              map[string]string `json:"abis"`
	TokenAddresses    map[string]string `json:"token_addresses"`
	OkuAddresses      map[string]string `json:"oku_addresses"`
	IonicAddresses    map[string]string `json:"ionic_addresses"`
	Endpoints         map[string]string `json:"enpoints"`
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
