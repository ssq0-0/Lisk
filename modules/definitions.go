package modules

import (
	"lisk/account"
	"lisk/globals"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type ModulesFasad interface {
	// Action(account *account.Account, amountKey string) error
	// ExecuteHardcodedTransaction(acc *account.Account) error
	Action(tokenIn, tokenOut common.Address, amountIn *big.Int, acc *account.Account, ta globals.ActionType) error
}
