package process

import (
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/models"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
)

func getMaxBalance(acc *account.Account, clients map[string]*ethClient.Client) (string, *big.Int, error) {
	var (
		maxChain string
		maxBal   = big.NewInt(0)
	)

	for chain, client := range clients {
		if chain == "lisk" {
			continue
		}

		balance, err := client.BalanceCheck(acc.Address, globals.WETH)
		if err != nil {
			return "", nil, fmt.Errorf("getMaxBalance: failed in chain %s: %w", chain, err)
		}

		if balance.Cmp(maxBal) > 0 {
			maxBal.Set(balance)
			maxChain = chain
		}
	}

	return maxChain, maxBal, nil
}

func validateSwapHistory(lastSwaps []models.SwapPair) error {
	if len(lastSwaps) > 1 {
		last := lastSwaps[len(lastSwaps)-1]
		prev := lastSwaps[len(lastSwaps)-2]
		if last.TokenFrom == prev.TokenFrom && last.TokenTo == prev.TokenTo {
			return fmt.Errorf("two identical swaps in a row not allowed")
		}
	}
	return nil
}

func selectValidSwapToken(tokenFrom common.Address, maxAttempts int) (common.Address, error) {
	for attempts := 0; attempts < maxAttempts; attempts++ {
		tokenTo := selectDifferentToken(tokenFrom)
		if !excludeSwap(tokenFrom, tokenTo) {
			return tokenTo, nil
		}
	}
	return common.Address{}, fmt.Errorf("не удалось выбрать допустимый токен для свопа после %d попыток", maxAttempts)
}

func selectDifferentToken(tokenFrom common.Address) common.Address {
	tokens := []common.Address{
		globals.LISK,
		globals.USDT,
		globals.USDC,
		globals.WETH,
	}

	for attempts := 0; attempts < 10; attempts++ {
		t := tokens[rand.Intn(len(tokens))]
		if t != tokenFrom {
			return t
		}
	}

	return globals.USDT
}

func excludeSwap(tokenFrom, tokenTo common.Address) bool {
	excludedPairs := map[common.Address]map[common.Address]bool{
		globals.USDT: {
			globals.LISK: true,
		},
		globals.USDC: {
			globals.WETH: true,
		},
		globals.WETH: {
			globals.USDC: true,
		},
	}

	if tos, exists := excludedPairs[tokenFrom]; exists {
		if _, excluded := tos[tokenTo]; excluded {
			return true
		}
	}
	return false
}

func canDoActionByBalance(token common.Address, acc *account.Account, client *ethClient.Client) (*big.Int, error) {
	balance, err := client.BalanceCheck(acc.Address, token)
	if err != nil {
		return nil, fmt.Errorf("canDoActionByBalance: balance check failed for %s: %w", acc.Address.Hex(), err)
	}

	if !checkMinimalAmount(balance, token) {
		return nil, fmt.Errorf("canDoActionByBalance: balance too low for token %s, account %s",
			token.Hex(), acc.Address.Hex())
	}

	return calculateAmount(balance, acc.BalancePercentUsage)
}

func checkMinimalAmount(balance *big.Int, token common.Address) bool {
	return balance.Cmp(globals.MinBalances[token]) >= 0
}

func calculateAmount(amount *big.Int, percent int) (*big.Int, error) {
	percentAmount := new(big.Int).Mul(amount, big.NewInt(int64(percent)))
	percentAmount.Div(percentAmount, big.NewInt(100))
	return percentAmount, nil
}

func updateLiquidityState(acc *account.Account, actionType globals.ActionType) {
	acc.LiquidityState.LastAction = actionType
	acc.LiquidityState.ActionCount++
}

func packActionProcessStruct(typeAction globals.ActionType, module string, amount *big.Int, tokenFrom, tokenTo common.Address) ActionProcess {
	return ActionProcess{
		TypeAction: typeAction,
		Module:     module,
		Amount:     amount,
		TokenFrom:  tokenFrom,
		TokenTo:    tokenTo,
	}
}
