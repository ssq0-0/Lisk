package models

import (
	"lisk/globals"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type SwapPair struct {
	TokenFrom common.Address
	TokenTo   common.Address
}

type LiquidityState struct {
	ActionCount               int
	LastAction                globals.ActionType
	PendingEnterAfterWithdraw bool
}

type FeePool struct {
	Fee          *big.Int
	PoolAddress  common.Address
	SqrtPriceX96 *big.Int
}

type RelayRequest struct {
	User                 string `json:"user"`
	OriginChainId        int    `json:"originChainId"`
	DestinationChainId   int    `json:"destinationChainId"`
	OriginCurrency       string `json:"originCurrency"`
	DestinationCurrency  string `json:"destinationCurrency"`
	Recipient            string `json:"recipient"`
	TradeType            string `json:"tradeType"`
	Amount               string `json:"amount"`
	Referrer             string `json:"referrer"`
	UseExternalLiquidity bool   `json:"useExternalLiquidity"`
	UseDepositAddress    bool   `json:"useDepositAddress"`
}

type RelayResponse struct {
	Steps []struct {
		Items []struct {
			Data struct {
				From    string `json:"from"`
				To      string `json:"to"`
				Data    string `json:"data"`
				Value   string `json:"value"`
				ChainID int    `json:"chainId"`
			} `json:"data"`
		} `json:"items"`
	} `json:"steps"`
}
