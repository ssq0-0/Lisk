package wraper

import (
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Wraper struct {
	Client *ethClient.Client
}

func NewWraper(client *ethClient.Client) (*Wraper, error) {
	return &Wraper{
		Client: client,
	}, nil
}

func (w *Wraper) Action(tokenIn, tokenOut common.Address, amountIn *big.Int, acc *account.Account, ta globals.ActionType) error {
	data, err := w.packData(ta, amountIn)
	if err != nil {
		return nil
	}

	value := amountIn
	if ta == globals.Unwrap {
		value = big.NewInt(0)
	}

	return w.Client.SendTransaction(acc.PrivateKey, acc.Address, globals.WETH, w.Client.GetNonce(acc.Address), value, data)
}

func (w *Wraper) packData(typePack globals.ActionType, amountIn *big.Int) ([]byte, error) {
	switch typePack {
	case globals.Wrap:
		return globals.Erc20ABI.Pack("deposit")
	case globals.Unwrap:
		return globals.Erc20ABI.Pack("withdraw", amountIn)
	default:
		return nil, fmt.Errorf("failed pack data for wrap/unwrap")
	}
}
