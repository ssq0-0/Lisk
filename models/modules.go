package models

import (
	"lisk/globals"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type SwapPair struct {
	TokenFrom common.Address
	TokenTo   common.Address
	Forced    bool
}

type WrapHistory struct {
	LastAction globals.ActionType
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

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type TaskResponse struct {
	Data struct {
		Userdrop struct {
			User struct {
				Rank     int    `json:"rank"`
				Points   int    `json:"points"`
				UpdateAt string `json:"updatedAt"`
			} `json:"user"`
			UpdateTaskStatus struct {
				Success  bool `json:"success"`
				Progress struct {
					IsCompleted bool   `json:"isCompleted"`
					CompletedAt string `json:"completedAt"`
				} `json:"progress"`
			} `json:"updateTaskStatus"`
		} `json:"userdrop"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type VersionInfo struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

type StatRecord struct {
	TotalSuccess int
	TodayDate    int // YYYYMMDD
	TodaySuccess int
}

type BlockscoutResp struct {
	GasPrice struct {
		Slow    float64 `json:"slow"`
		Average float64 `json:"average"`
		Fast    float64 `json:"fast"`
	} `json:"gas_prices"`
}
