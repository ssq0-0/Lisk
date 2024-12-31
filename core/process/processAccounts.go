package process

import (
	"context"
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
	"golang.org/x/sync/errgroup"
)

type ActionProcess struct {
	TokenFrom  common.Address
	TokenTo    common.Address
	Amount     *big.Int
	TypeAction globals.ActionType
	Module     string
}

var actionGenerators = map[string]func(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error){
	"Oku":   generateSwap,
	"Ionic": generateLiquidity,
	"Relay": generateBridgeToLisk,
}

func ProcessAccounts(accs []*account.Account, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client) error {
	if len(accs) == 0 || len(mod) == 0 || len(clients) == 0 {
		return fmt.Errorf("One of the elements is missing(accounts, modules, eth clients). Check the settings.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	const maxConcurrency = 10
	semaphore := make(chan struct{}, maxConcurrency)

	var (
		mu   sync.Mutex
		errs []error
	)

	for _, acc := range accs {
		acc := acc

		g.Go(func() error {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err := processSingleAccount(ctx, acc, mod, clients); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}

			return nil
		})

	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("ProcessAccount interrupted: %w", err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("ProcessAccount encountered %d errors", len(errs))
	}

	return nil
}

func processSingleAccount(ctx context.Context, acc *account.Account, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client) error {
	if err := performActions(acc, mod, clients); err != nil {
		logger.GlobalLogger.Errorf("failed to perform actions for account %s: %v", acc.Address.Hex(), err)
		return fmt.Errorf("account %s: performActions error: %w", acc.Address.Hex(), err)
	}

	logger.GlobalLogger.Infof("The job is done. Account statistics: wallet: %s Success actions: %d",
		acc.Address.Hex(), acc.Stats)

	return nil
}

func performActions(acc *account.Account, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client) error {
	availableModules := make([]string, 0, len(mod))
	for moduleName := range mod {
		availableModules = append(availableModules, moduleName)
	}

	times := generateTimeWindow(acc.ActionsTime, acc.ActionsCount)

	totalActions := acc.ActionsCount

	for i := 0; i < totalActions; i++ {
		if i >= len(times) {
			newTime := generateTimeWindow(acc.ActionsTime, acc.ActionsCount)[0]
			times = append(times, newTime)
		}

		action, err := generateNextAction(acc, availableModules, clients)
		if err != nil {
			totalActions++
			continue
		}

		logger.GlobalLogger.Infof("Processing action for %s: %v, type: %v", acc.Address.Hex(), action.Module, action.TypeAction)

		moduleFasad, exists := mod[action.Module]
		if !exists || moduleFasad == nil {
			logger.GlobalLogger.Warnf("Module '%s' not found for action. Will add a new attempt.", action.Module)
			totalActions++
			continue
		}

		if err := moduleFasad.Action(action.TokenFrom, action.TokenTo, action.Amount, acc, action.TypeAction); err != nil {
			logger.GlobalLogger.Warnf("Failed to perform action: %v. Adding a new attempt.", err)
			totalActions++
			times = append(times, generateTimeWindow(acc.ActionsTime, acc.ActionsCount)[0])
			continue
		}

		acc.Stats += 1
		logger.GlobalLogger.Infof("The action for account %v has been completed. Sleep %v before the next action.", acc.Address.Hex(), times[i])
		time.Sleep(times[i])
	}

	return nil
}

func generateBridgeToLisk(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	chain, balance, err := getMaxBalance(acc, clients)
	if err != nil {
		return ActionProcess{TypeAction: globals.Unknown}, fmt.Errorf("failed get max balance in all chains: %v", err)
	}

	if chain == "" {
		return ActionProcess{TypeAction: globals.Unknown}, fmt.Errorf("no balance in other networks")
	}
	logger.GlobalLogger.Infof("Bridge to LISK. From: %s.", chain)
	return bridgeToLisk(acc, balance, chain, clients[chain])
}

func bridgeToLisk(acc *account.Account, balance *big.Int, chain string, client *ethClient.Client) (ActionProcess, error) {
	percentBalance := new(big.Int).Mul(balance, big.NewInt(70))
	percentBalance.Div(percentBalance, big.NewInt(100))

	return ActionProcess{
		TokenFrom:  globals.NATIVE,
		TypeAction: globals.Bridge[chain],
		Amount:     percentBalance,
		Module:     "Relay",
	}, nil
}

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

func generateNextAction(acc *account.Account, availableModules []string, clients map[string]*ethClient.Client) (ActionProcess, error) {
	if len(availableModules) == 0 {
		return ActionProcess{TypeAction: globals.Unknown}, fmt.Errorf("no available modules for account %s", acc.Address.Hex())
	}

	selectedModule := availableModules[rand.Intn(len(availableModules))]

	generator, exists := actionGenerators[selectedModule]
	if !exists {
		return ActionProcess{TypeAction: globals.Unknown}, fmt.Errorf("no action generator for module '%s'", selectedModule)
	}

	return generator(acc, clients)
}

func generateSwap(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
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
	const maxAttempts = 5
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

	amount, err := canDoActionByBalance(tokenFrom, acc, clients["lisk"])
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

func generateLiquidity(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
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
