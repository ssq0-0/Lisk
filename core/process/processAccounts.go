package process

import (
	"context"
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/logger"
	"lisk/modules"
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

func ProcessAccounts(accs []*account.Account, selectModule string, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client) error {
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

			if err := processSingleAccount(ctx, acc, selectModule, mod, clients); err != nil {
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

	if err := WriteWeeklyStats(accs); err != nil {
		return fmt.Errorf("failed to write weekly stats: %w", err)
	}

	return nil
}

func processSingleAccount(ctx context.Context, acc *account.Account, selectModule string, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client) error {
	if err := performActions(acc, selectModule, mod, clients); err != nil {
		logger.GlobalLogger.Errorf("failed to perform actions for account %s: %v", acc.Address.Hex(), err)
		return fmt.Errorf("account %s: performActions error: %w", acc.Address.Hex(), err)
	}

	logger.GlobalLogger.Infof("The %s is done", acc.Address.Hex())

	return nil
}

func performActions(acc *account.Account, selectModule string, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client) error {
	totalActions, exists := globals.LimitedModules[selectModule]
	if !exists {
		totalActions = acc.ActionsCount
	}

	const maxRetriesPerAction = 3

	var successfulActions int

	for successfulActions < totalActions {
		sleepDuration := generateTimeWindow(acc.ActionsTime, acc.ActionsCount)[0]

		action, err := generateNextAction(acc, selectModule, clients)
		if err != nil {
			logger.GlobalLogger.Warnf("Cannot generate action: %v", err)
			continue
		}

		retryCount := 0
	actionLoop:
		for {
			moduleFasad, exists := mod[action.Module]
			if !exists || moduleFasad == nil {
				logger.GlobalLogger.Warnf("Module '%s' not found. Skipping action.", action.Module)
				break actionLoop
			}

			if err := moduleFasad.Action(action.TokenFrom, action.TokenTo, action.Amount, acc, action.TypeAction); err != nil {
				retryCount++
				if retryCount <= maxRetriesPerAction {
					logger.GlobalLogger.Warnf("Action failed: %v. Retry %d of %d...", err, retryCount, maxRetriesPerAction)
					time.Sleep(5 * time.Second)
					continue
				} else {
					logger.GlobalLogger.Errorf("Action failed after %d retries. Skip action.", maxRetriesPerAction)
					break actionLoop
				}
			}

			acc.Stats[selectModule]++
			successfulActions++

			logger.GlobalLogger.Infof("Action for account %v has been completed. Sleep %v", acc.Address.Hex(), sleepDuration)
			time.Sleep(sleepDuration)

			break actionLoop
		}
	}

	return nil
}
