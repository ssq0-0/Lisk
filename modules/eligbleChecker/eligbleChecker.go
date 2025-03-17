package eligbleChecker

import (
	"lisk/account"
	"lisk/globals"
	"lisk/httpClient"
	"lisk/models"
	"lisk/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Checker struct {
	HttpClinet *httpClient.HttpClient
}

func NewChecker(hc *httpClient.HttpClient) (*Checker, error) {
	return &Checker{
		HttpClinet: hc,
	}, nil
}

func (c *Checker) Action(tokenIn, tokenOut common.Address, amountIn *big.Int, acc *account.Account, operation globals.ActionType) error {
	requstPayload := map[string]string{
		"wallet_address": acc.Address.String(),
	}

	resp := &models.CheckResponse{} // Создаем объект перед использованием
	if err := c.HttpClinet.SendJSONRequest("https://lisk.com/wp-json/bornfight/v1/eligibility-check", "POST", requstPayload, resp); err != nil {
		return err
	}

	return utils.WriteToCSV(utils.GetPath("eligble"), acc.Address.Hex(), resp.Message)
}
