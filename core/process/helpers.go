package process

import (
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/models"
	"lisk/modules"
	"math/big"
	"math/rand"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

func validateInputData(accs []*account.Account, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client) error {
	if len(accs) == 0 || len(mod) == 0 || len(clients) == 0 {
		return fmt.Errorf("One of the elements is missing(accounts, modules, eth clients). Check the settings.")
	}
	return nil
}

func determineActionCount(acc *account.Account, selectModule string) int {
	totalActions, exists := globals.LimitedModules[selectModule]
	if !exists {
		totalActions = acc.ActionsCount
	}

	return totalActions
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

func selectTokenFrom(acc *account.Account) common.Address {
	if len(acc.LastSwaps) == 0 {
		return globals.WETH
	}

	lastTokenTo := acc.LastSwaps[len(acc.LastSwaps)-1].TokenTo
	if lastTokenTo != globals.LISK {
		return lastTokenTo
	}
	return globals.WETH
}

func selectDifferentToken(token common.Address) common.Address {
	tokens := []common.Address{
		globals.LISK,
		globals.USDT,
		globals.USDC,
		globals.WETH,
	}

	for attempts := 0; attempts < 10; attempts++ {
		t := tokens[rand.Intn(len(tokens))]
		if t != token && !excludeSwap(token, t) {
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
			globals.LISK: true,
		},
		globals.WETH: {
			globals.USDC: true,
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
	minBalance, exists := globals.MinBalances[token]
	if !exists {
		return false
	}
	return balance.Cmp(minBalance) >= 0
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

func calculateAmount(amount *big.Int, percent int) (*big.Int, error) {
	percentAmount := new(big.Int).Mul(amount, big.NewInt(int64(percent)))
	percentAmount.Div(percentAmount, big.NewInt(100))
	return percentAmount, nil
}

func updateSwapHistory(acc *account.Account, tokenFrom, tokenTo common.Address) {
	acc.LastSwaps = append(acc.LastSwaps, models.SwapPair{
		TokenFrom: tokenFrom,
		TokenTo:   tokenTo,
	})
}

func updateLiquidityState(acc *account.Account, actionType globals.ActionType) {
	acc.LiquidityState.LastAction = actionType
	acc.LiquidityState.ActionCount++
}

func validateNativeBalance(addr common.Address, client *ethClient.Client) (*big.Int, error) {
	balance, err := client.BalanceCheck(addr, globals.WETH)
	if err != nil {
		return big.NewInt(0), err
	}

	if balance.Cmp(globals.MinETHForTx) < 0 {
		return balance, fmt.Errorf("Native(ETH) balance too low")
	}

	return balance, nil
}

func isCriticalError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	errorSubstrings := []string{
		"balance too low",
		"Gas wait timeout has been exceeded",
	}

	for _, substr := range errorSubstrings {
		if strings.Contains(errMsg, substr) {
			return true
		}
	}
	return false
}
