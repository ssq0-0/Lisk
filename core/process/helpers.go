package process

import (
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func generateTimeWindow(totalTime, actionCount int) []time.Duration {
	if actionCount <= 0 {
		return nil
	}

	baseInterval := time.Duration(totalTime) * time.Minute / time.Duration(actionCount)
	const variationFactor = 0.2
	const minVariation = time.Second
	const maxVariation = 30 * time.Second

	intervals := make([]time.Duration, 0, actionCount)
	for i := 0; i < actionCount; i++ {
		variation := float64(baseInterval) * variationFactor

		if variation < float64(minVariation) {
			variation = float64(minVariation)
		} else if variation > float64(maxVariation) {
			variation = float64(maxVariation)
		}

		randomVariation := time.Duration(rand.Float64()*2*variation - variation)
		interval := baseInterval + randomVariation

		if interval < 0 {
			interval = baseInterval
		}

		intervals = append(intervals, interval)
	}

	return intervals
}

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

func generateNextAction(acc *account.Account, selectedModule string, clients map[string]*ethClient.Client) (ActionProcess, error) {
	generator, exists := actionGenerators[selectedModule]
	if !exists {
		return ActionProcess{TypeAction: globals.Unknown}, fmt.Errorf("no action generator for module '%s'", selectedModule)
	}

	return generator(acc, clients)
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
