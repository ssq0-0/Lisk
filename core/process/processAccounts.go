package process

import (
	"context"
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/logger"
	"lisk/modules"
	"lisk/utils"
	"math/big"
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

func ProcessAccounts(accs []*account.Account, selectModule string, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client, memory *Memory) error {
	if err := validateInputData(accs, mod, clients); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	const maxConcurrency = 10
	semaphore := make(chan struct{}, maxConcurrency)

	var (
		mu       sync.Mutex
		nonFatal []error
	)

	for _, acc := range accs {
		currentAcc := acc

		g.Go(func() error {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err := processSingleAccount(ctx, currentAcc, selectModule, mod, clients, memory); err != nil {
				mu.Lock()
				nonFatal = append(nonFatal, err)
				mu.Unlock()
			}

			return nil
		})

	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("ProcessAccount interrupted: %w", err)
	}

	if len(nonFatal) > 0 {
		return fmt.Errorf("ProcessAccount encountered %d errors", len(nonFatal))
	}

	if err := WriteWeeklyStats(accs); err != nil {
		return fmt.Errorf("failed to write weekly stats: %w", err)
	}

	return nil
}

func processSingleAccount(ctx context.Context, acc *account.Account, selectModule string, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client, memory *Memory) error {
	if err := performActions(acc, selectModule, mod, clients, memory); err != nil {
		logger.GlobalLogger.Errorf("[%v] failed to perform actions: %v", acc.Address.Hex(), err)
		return fmt.Errorf("[%v] performActions error: %w", acc.Address.Hex(), err)
	}

	logger.GlobalLogger.Infof("[%v] Processing is done", acc.Address.Hex())

	return nil
}

func performActions(acc *account.Account, selectModule string, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client, memory *Memory) error {
	if _, err := validateNativeBalance(acc.Address, clients["lisk"]); err != nil {
		if isCriticalError(err) && (selectModule != "Checker" && selectModule != "Portal_daily_check" && selectModule != "Portal_main_tasks" && selectModule != "BalanceCheck") {
			logger.GlobalLogger.Warnf("[%v] Insufficient ETH  balance. Stop trying.", acc.Address.Hex())
			if err := utils.ReplacePrivateKey(acc.RawPK, acc.Address.Hex()); err != nil {
				logger.GlobalLogger.Errorf("[%v] Failed to replace private key: %v", acc.Address, err)
				return err
			}
			return err
		}
	}

	state, err := memory.LoadState(acc.Address.Hex())
	if err != nil {
		return fmt.Errorf("failed to load state for account %s: %v", acc.Address.Hex(), err)
	}

	const maxRetriesPerAction = 3
	var successfulActions int
	if state != nil {
		successfulActions = state.LastActionIndex
		selectModule = state.Module
		logger.GlobalLogger.Infof("[%s] Resuming from action index %d", acc.Address.Hex(), successfulActions)
	}

	totalActions := determineActionCount(acc, selectModule)

	for successfulActions < totalActions {
		sleepDuration := generateTimeWindow(acc.ActionsTime, totalActions)[0]

		action, err := generateNextAction(acc, selectModule, clients)
		if err != nil {
			if isCriticalError(err) {
				logger.GlobalLogger.Warnf("[%v] Insufficient balance for swap. Stop trying.", acc.Address.Hex())
				return err
			}
			logger.GlobalLogger.Warnf("[%v] Cannot generate action: %v", acc.Address, err)
			continue
		}

		retryCount := 0
	actionLoop:
		for {
			moduleFasad, exists := mod[action.Module]
			if !exists || moduleFasad == nil {
				logger.GlobalLogger.Warnf("[%v] Module '%s' not found. Skipping action for %s.", acc.Address, action.Module, acc.Address)
				break actionLoop
			}

			if err := moduleFasad.Action(action.TokenFrom, action.TokenTo, action.Amount, acc, action.TypeAction); err != nil {
				if isCriticalError(err) {
					logger.GlobalLogger.Warnf("[%v] Insufficient balance for swap. Stop trying.", acc.Address.Hex())
					return err
				}

				retryCount++
				if retryCount <= maxRetriesPerAction {
					logger.GlobalLogger.Warnf("[%v] Action failed for: %v. Retry %d of %d...", acc.Address, err, retryCount, maxRetriesPerAction)
					time.Sleep(5 * time.Second)
					continue
				} else {
					logger.GlobalLogger.Errorf("[%v] Action failed after %d retries. Skip action.", acc.Address, maxRetriesPerAction)
					break actionLoop
				}
			}

			if err := memory.UpdateState(acc.Address.Hex(), selectModule, successfulActions+1); err != nil {
				logger.GlobalLogger.Warnf("[%s] Failed to update state: %v", acc.Address.Hex(), err)
			}
			acc.Stats[selectModule]++
			successfulActions++

			logger.GlobalLogger.Infof("[%v] Action has been completed. Sleep %v", acc.Address.Hex(), sleepDuration)
			time.Sleep(sleepDuration)

			break actionLoop
		}
	}

	if successfulActions == totalActions {
		if err := memory.ClearState(acc.Address.Hex()); err != nil {
			logger.GlobalLogger.Warnf("[%s] Failed to clear state: %v", acc.Address.Hex(), err)
		}
		logger.GlobalLogger.Infof("[%s] All actions completed. State cleared.", acc.Address.Hex())
	}
	return nil
}
