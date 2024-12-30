package process

import (
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/logger"
	"lisk/models"
	"lisk/modules"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type ActionProcess struct {
	TokenFrom  common.Address
	TokenTo    common.Address
	Amount     *big.Int
	TypeAction globals.ActionType
	Module     string
}

func ProcessAccount(accs []*account.Account, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client) error {
	var wg sync.WaitGroup

	for _, acc := range accs {
		wg.Add(1)

		go func(acc *account.Account) {
			defer wg.Done()

			needsBridged, err := shouldBridge(acc, clients)
			if err != nil {
				logger.GlobalLogger.Error(err)
			}

			if needsBridged {
				if err := checkAndBridgeToLisk(acc, mod, clients); err != nil {
					logger.GlobalLogger.Errorf("failed bridge to lisk: %v", err)
				} else {
					logger.GlobalLogger.Infof("successful bridge to Lisk. Sleep 3 minute.")
					time.Sleep(time.Minute * 3)
				}
			}

			if err := performActions(acc, mod, clients); err != nil {
				logger.GlobalLogger.Errorf("failed to perform actions for account %s: %v", acc.Address.Hex(), err)
			}

		}(acc)
	}

	wg.Wait()
	return nil
}

func checkAndBridgeToLisk(acc *account.Account, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client) error {
	chain, balance, err := getMaxBalance(acc, clients)
	if err != nil {
		return fmt.Errorf("failed get max balance in all chains: %v", err)
	}

	logger.GlobalLogger.Infof("Bridge to LISK. From: %s.", chain)
	if err := bridgeToLisk(acc, balance, chain, clients[chain], mod); err != nil {
		return err
	}
	return nil
}

func performActions(acc *account.Account, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client) error {
	times := generateTimeWindow(acc.ActionsTime, acc.ActionsCount)

	actionsCount := acc.ActionsCount
	for i := 0; i < actionsCount; i++ {
		if i >= len(times) {
			newTime := generateTimeWindow(acc.ActionsTime, 1)[0]
			times = append(times, newTime)
		}

		action, err := generateNextAction(acc, clients["lisk"])
		if err != nil {
			logger.GlobalLogger.Warnf("failed to generate action. Will add a new attempt. err: %v", err)
			actionsCount++
			continue
		}

		logger.GlobalLogger.Infof("Processing action for %s: %v, type: %v", acc.Address.Hex(), action.Module, action.TypeAction)

		moduleFasad, exists := mod[action.Module]
		if !exists || moduleFasad == nil {
			logger.GlobalLogger.Warnf("Module '%s' not found for action. Will add a new attempt.", action.Module)
			actionsCount++
			continue
		}

		if err := moduleFasad.Action(action.TokenFrom, action.TokenTo, action.Amount, acc, action.TypeAction); err != nil {
			logger.GlobalLogger.Warnf("Failed to perform action: %v. Adding a new attempt.", err)
			actionsCount++
			times = append(times, generateTimeWindow(acc.ActionsTime, acc.ActionsCount)[0])
			continue
		}

		logger.GlobalLogger.Infof("Действие %d для аккаунта %v выполнено. Спим %v перед следующим действием.",
			i, acc.Address.Hex(), times[i])
		time.Sleep(times[i])
	}

	return nil
}

func shouldBridge(acc *account.Account, clients map[string]*ethClient.Client) (bool, error) {
	liskBalance, err := clients["lisk"].BalanceCheck(acc.Address, globals.WETH)
	if err != nil {
		return false, err
	}

	if liskBalance.Cmp(globals.MinBalances[globals.WETH]) < 0 {
		return true, nil
	}

	return false, nil
}

func bridgeToLisk(acc *account.Account, balance *big.Int, chain string, client *ethClient.Client, mod map[string]modules.ModulesFasad) error {
	percentBalance := new(big.Int).Mul(balance, big.NewInt(70))
	percentBalance.Div(percentBalance, big.NewInt(100))

	if err := mod["Relay"].Action(globals.NATIVE, globals.NATIVE, percentBalance, acc, globals.Bridge[chain]); err != nil {
		return err
	}

	return nil
}

func generateTimeWindow(totalTime, actionCount int) []time.Duration {
	if actionCount <= 0 {
		return nil
	}

	baseInterval := time.Duration(totalTime) * time.Minute / time.Duration(actionCount)
	variationFactor := 0.2

	minVariation := time.Second
	maxVariation := 30 * time.Second

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
			return "", nil, fmt.Errorf("failed to get balance in chain %s: %v", chain, err)
		}

		if balance.Cmp(maxBal) > 0 {
			maxBal.Set(balance)
			maxChain = chain
		}
	}

	return maxChain, maxBal, nil
}

func generateNextAction(acc *account.Account, client *ethClient.Client) (ActionProcess, error) {
	if len(acc.LastSwaps) == 0 && acc.LiquidityState.ActionCount == 0 {
		return generateSwap(acc, client)
	}

	if rand.Intn(2) == 0 {
		action, err := generateSwap(acc, client)
		if err != nil {
			return generateLiquidity(acc, client)
		}
		return action, nil
	} else {
		action, err := generateLiquidity(acc, client)
		if err != nil {
			return generateSwap(acc, client)
		}
		return action, nil
	}
}

func generateSwap(acc *account.Account, client *ethClient.Client) (ActionProcess, error) {
	if len(acc.LastSwaps) > 1 {
		last := acc.LastSwaps[len(acc.LastSwaps)-1]
		prev := acc.LastSwaps[len(acc.LastSwaps)-2]
		if last.TokenFrom == prev.TokenFrom && last.TokenTo == prev.TokenTo {
			return ActionProcess{TypeAction: globals.Unknown},
				fmt.Errorf("two identical swaps in a row not allowed")
		}
	}

	var tokenFrom common.Address
	if len(acc.LastSwaps) == 0 {
		tokenFrom = globals.WETH
	} else {
		lastTokenTo := acc.LastSwaps[len(acc.LastSwaps)-1].TokenTo
		if lastTokenTo != globals.LISK {
			tokenFrom = lastTokenTo
		} else {
			tokenFrom = globals.WETH
		}
	}

	var tokenTo common.Address
	maxAttempts := 5
	for attempts := 0; attempts < maxAttempts; attempts++ {
		tokenTo = selectDifferentToken(tokenFrom)
		if !excludeSwap(tokenFrom, tokenTo) {
			break
		}
		if attempts == maxAttempts-1 {
			return ActionProcess{TypeAction: globals.Unknown}, fmt.Errorf("failed to select non-excluded token after %d attempts", maxAttempts)
		}
	}
	acc.LastSwaps = append(acc.LastSwaps, models.SwapPair{
		TokenFrom: tokenFrom,
		TokenTo:   tokenTo,
	})

	amount, err := canDoActionByBalance(tokenFrom, acc, client)
	if err != nil {
		return ActionProcess{TypeAction: globals.Unknown}, err
	}

	return ActionProcess{
		TokenFrom:  tokenFrom,
		TokenTo:    tokenTo,
		Amount:     amount,
		TypeAction: globals.Swap,
		Module:     "Oku",
	}, nil
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

func generateLiquidity(acc *account.Account, client *ethClient.Client) (ActionProcess, error) {
	liqState := acc.LiquidityState

	if liqState.ActionCount == 0 {
		action := ActionProcess{
			TokenFrom:  globals.USDT,
			Amount:     big.NewInt(9e4),
			TypeAction: globals.Supply,
			Module:     "Ionic",
		}
		updateLiquidityState(acc, action.TypeAction)
		return action, nil
	}

	if liqState.ActionCount == 1 {
		action := ActionProcess{
			TypeAction: globals.EnterMarket,
			Module:     "Ionic",
		}
		updateLiquidityState(acc, action.TypeAction)
		return action, nil
	}

	if liqState.LastAction == globals.Redeem {
		if !liqState.PendingEnterAfterWithdraw {
			liqState.PendingEnterAfterWithdraw = true
			action := ActionProcess{
				TokenFrom:  globals.USDT,
				Amount:     big.NewInt(1e4),
				TypeAction: globals.Supply,
				Module:     "Ionic",
			}
			updateLiquidityState(acc, action.TypeAction)
			return action, nil
		} else {
			liqState.PendingEnterAfterWithdraw = false
			action := ActionProcess{
				TypeAction: globals.EnterMarket,
				Module:     "Ionic",
			}
			updateLiquidityState(acc, action.TypeAction)
			return action, nil
		}
	}

	if liqState.LastAction == globals.Borrow {
		action := ActionProcess{
			TokenFrom:  globals.LISK,
			Amount:     big.NewInt(1e2),
			TypeAction: globals.Repay,
			Module:     "Ionic",
		}
		updateLiquidityState(acc, action.TypeAction)
		return action, nil
	}

	if liqState.LastAction == globals.Repay || liqState.LastAction == globals.EnterMarket {
		action := ActionProcess{
			TokenFrom:  globals.LISK,
			Amount:     big.NewInt(1e2),
			TypeAction: globals.Borrow,
			Module:     "Ionic",
		}
		updateLiquidityState(acc, action.TypeAction)
		return action, nil
	}

	if liqState.LastAction == globals.Supply {
		action := ActionProcess{
			TokenFrom:  globals.USDT,
			Amount:     big.NewInt(1e4),
			TypeAction: globals.Supply,
			Module:     "Ionic",
		}
		updateLiquidityState(acc, action.TypeAction)
		return action, nil
	}

	return ActionProcess{TypeAction: globals.Unknown}, nil
}

func canDoActionByBalance(token common.Address, acc *account.Account, client *ethClient.Client) (*big.Int, error) {
	balance, err := client.BalanceCheck(acc.Address, token)
	if err != nil {
		return nil, err
	}

	if !checkMinimalAmount(balance, token) {
		return nil, fmt.Errorf("balance is low. Rollback.")
	}

	return calculateAmount(balance, acc.BalancePercentUsage)
}

func checkMinimalAmount(balance *big.Int, token common.Address) bool {
	if balance.Cmp(globals.MinBalances[token]) < 0 {
		return false
	}

	return true
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
