package balanceChecker

import (
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Checker struct {
	Client *ethClient.Client
	Tokens map[string]common.Address
}

func NewChecker(client *ethClient.Client, tokens map[string]common.Address) (*Checker, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("0 tokens for balance check.")
	}

	return &Checker{
		Client: client,
		Tokens: tokens,
	}, nil
}

func (c *Checker) Action(tokenIn, tokenOut common.Address, amountIn *big.Int, acc *account.Account, ta globals.ActionType) error {
	balances := make(map[string]map[string]string)

	for token, address := range c.Tokens {
		balance, err := c.Client.BalanceCheck(acc.Address, address)
		if err != nil {
			return fmt.Errorf("failed to check balance for token %s: %w", token, err)
		}

		result := utils.ConvertFromWei(balance, globals.DecimalsMap[address])
		if _, exists := balances[acc.Address.Hex()]; !exists {
			balances[acc.Address.Hex()] = make(map[string]string)
		}

		balances[acc.Address.Hex()][token] = result
	}

	if err := utils.WriteAddrTokenBalancesToFile(balances); err != nil {
		return fmt.Errorf("failed to write balances to file: %w", err)
	}

	return nil
}
