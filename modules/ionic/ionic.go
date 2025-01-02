package ionic

import (
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Ionic struct {
	ABI    *abi.ABI
	Client *ethClient.Client
	Tokens map[common.Address]common.Address
}

func NewIonic(addresses map[string]string, abi *abi.ABI, client *ethClient.Client) (*Ionic, error) {
	if addresses == nil {
		return nil, fmt.Errorf("addresses map cannot be nil")
	}

	tokens := make(map[common.Address]common.Address)
	for token, address := range addresses {
		tokenAddr := common.HexToAddress(token)
		contractAddr := common.HexToAddress(address)
		tokens[tokenAddr] = contractAddr
	}

	return &Ionic{
		ABI:    abi,
		Tokens: tokens,
		Client: client,
	}, nil
}

func (i *Ionic) Action(tokenIn, tokenOut common.Address, amountIn *big.Int, acc *account.Account, operation globals.ActionType) error {
	data, err := i.prepareTx(operation, amountIn, tokenIn)
	if err != nil {
		return err
	}

	switch operation {
	case globals.Supply, globals.Repay:
		if err := i.ensureAllowance(tokenIn, acc, amountIn); err != nil {
			return fmt.Errorf("failed to approve tokens: %w", err)
		}
	}
	addressCA := i.prepareCA(tokenIn, operation)

	return i.Client.SendTransaction(acc.PrivateKey, acc.Address, addressCA, i.Client.GetNonce(acc.Address), big.NewInt(0), data)
}

func (i *Ionic) ensureAllowance(tokenIn common.Address, acc *account.Account, amountIn *big.Int) error {
	if _, err := i.Client.ApproveTx(tokenIn, i.Tokens[tokenIn], acc, globals.MaxRepayBigInt, false); err != nil {
		return err
	}

	return nil
}

func (i *Ionic) prepareTx(operation globals.ActionType, amountIn *big.Int, token common.Address) ([]byte, error) {
	switch operation {
	case globals.Supply:
		return i.ABI.Pack("mint", amountIn)
	case globals.Redeem:
		return i.ABI.Pack("redeemUnderlying", amountIn)
	case globals.Borrow:
		return i.ABI.Pack("borrow", amountIn)
	case globals.Repay:
		return i.ABI.Pack("repayBorrow", amountIn)
	case globals.EnterMarket:
		return i.ABI.Pack("enterMarkets", []common.Address{common.HexToAddress("0x0D72f18BC4b4A2F0370Af6D799045595d806636F")})
	case globals.ExitMarket:
		return i.ABI.Pack("exitMarket", i.Tokens[token])
	default:
		return nil, fmt.Errorf("unknow operation in Ionic")
	}
}

func (i *Ionic) prepareCA(token common.Address, operation globals.ActionType) common.Address {
	switch operation {
	case globals.EnterMarket, globals.ExitMarket:
		return common.HexToAddress("0xF448A36feFb223B8E46e36FF12091baBa97bdF60")
	default:
		return i.Tokens[token]
	}
}
