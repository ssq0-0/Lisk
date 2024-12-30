package test

import (
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/logger"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type IonicTest struct {
	ABI    *abi.ABI
	Client *ethClient.Client
	Tokens map[common.Address]common.Address
}

func NewTestIonic(addresses map[string]string, abi *abi.ABI, client *ethClient.Client) (*IonicTest, error) {
	if addresses == nil {
		return nil, fmt.Errorf("addresses map cannot be nil")
	}

	tokens := make(map[common.Address]common.Address)
	for token, address := range addresses {
		tokenAddr := common.HexToAddress(token)
		contractAddr := common.HexToAddress(address)
		tokens[tokenAddr] = contractAddr
	}

	logger.GlobalLogger.Infof("инициализация токенов: %v", tokens)
	return &IonicTest{
		ABI:    abi,
		Tokens: tokens,
		Client: client,
	}, nil
}

func (i *IonicTest) Action(tokenIn, tokenOut common.Address, amountIn *big.Int, acc *account.Account, ta globals.ActionType) error {
	logger.GlobalLogger.Infof("это входная функция ионика. Вызов...")
	_, _ = i.prepareTx(ta, amountIn, tokenIn)
	return nil
}

func (i *IonicTest) prepareTx(operation globals.ActionType, amountIn *big.Int, token common.Address) ([]byte, error) {
	switch operation {
	case globals.Supply:
		logger.GlobalLogger.Infof("суплай, параметры: сумма %v, токен %v", amountIn, token)
		return i.ABI.Pack("mint", amountIn)
	case globals.Redeem:
		logger.GlobalLogger.Infof("редем, параметры:сумма %v, токен %v", amountIn, token)
		return i.ABI.Pack("redeemUnderlying", amountIn)
	case globals.Borrow:
		logger.GlobalLogger.Infof("борроу, параметры: сумма %v, токен %v", amountIn, token)
		return i.ABI.Pack("borrow", amountIn)
	case globals.Repay:
		logger.GlobalLogger.Infof("репей, параметры: сумма %v, токен %v", amountIn, token)
		return i.ABI.Pack("repayBorrow", amountIn)
	case globals.EnterMarket:
		logger.GlobalLogger.Infof("ентермаркет, параметры: сумма %v, токен %v", amountIn, token)
		return i.ABI.Pack("enterMarkets", []common.Address{token})
	case globals.ExitMarket:
		logger.GlobalLogger.Infof("ексит маркет, параметры:сумма %v, токен %v", amountIn, token)
		return i.ABI.Pack("exitMarket", token)
	default:
		return nil, fmt.Errorf("unknow operation in Ionic")
	}
}
