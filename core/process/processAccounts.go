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

	return nil
}

func processSingleAccount(ctx context.Context, acc *account.Account, selectModule string, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client) error {
	if err := performActions(acc, selectModule, mod, clients); err != nil {
		logger.GlobalLogger.Errorf("failed to perform actions for account %s: %v", acc.Address.Hex(), err)
		return fmt.Errorf("account %s: performActions error: %w", acc.Address.Hex(), err)
	}

	logger.GlobalLogger.Infof("The job is done. Account statistics: wallet: %s Success actions: %d", acc.Address.Hex(), acc.Stats)

	return nil
}

func performActions(acc *account.Account, selectModule string, mod map[string]modules.ModulesFasad, clients map[string]*ethClient.Client) error {
	times := generateTimeWindow(acc.ActionsTime, acc.ActionsCount)
	totalActions := acc.ActionsCount

	if globals.LimitedModules[selectModule] {
		totalActions = 1
	}

	for i := 0; i < totalActions; i++ {
		if i >= len(times) {
			newTime := generateTimeWindow(acc.ActionsTime, acc.ActionsCount)[0]
			times = append(times, newTime)
		}

		action, err := generateNextAction(acc, selectModule, clients)
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
			logger.GlobalLogger.Warnf("Failed to perform action: %v. Adding a new attempt and Sleep 15 seconds.", err)
			totalActions++
			times = append(times, generateTimeWindow(acc.ActionsTime, acc.ActionsCount)[0])
			time.Sleep(15 * time.Second)
			continue
		}

		acc.Stats += 1
		logger.GlobalLogger.Infof("The action for account %v has been completed. Sleep %v before the next action.", acc.Address.Hex(), times[i])
		time.Sleep(times[i])
	}

	return nil
}
