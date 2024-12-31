package modules

import (
	"context"
	"errors"
	"fmt"
	"lisk/config"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/modules/dex"
	"lisk/modules/ionic"
	"lisk/modules/relay"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"golang.org/x/sync/errgroup"
)

type ModuleFactory func(cfg *config.Config, clients map[string]*ethClient.Client) (ModulesFasad, error)

func ModulesInit(cfg *config.Config, selectModules string, abis map[string]*abi.ABI, clients map[string]*ethClient.Client) (map[string]ModulesFasad, error) {
	if strings.Trim(selectModules, "") == "" {
		return nil, errors.New("no modules to initialize")
	}

	modules := make(map[string]ModuleFactory)

	allModules := map[string]ModuleFactory{
		"Oku": func(cfg *config.Config, clients map[string]*ethClient.Client) (ModulesFasad, error) {
			return dex.NewDex(cfg.OkuAddresses, abis["oku"], clients["lisk"])
		},
		"Ionic": func(cfg *config.Config, clients map[string]*ethClient.Client) (ModulesFasad, error) {
			return ionic.NewIonic(cfg.IonicAddresses, abis["ionic"], clients["lisk"])
		},
		"Relay": func(cfg *config.Config, clients map[string]*ethClient.Client) (ModulesFasad, error) {
			relayClients := map[globals.ActionType]*ethClient.Client{
				globals.LineaBridge:    clients["linea"],
				globals.ArbitrumBridge: clients["arbitrum"],
				globals.OptimismBridge: clients["optimism"],
				globals.BaseBridge:     clients["base"],
			}

			return relay.NewRelay(relayClients, cfg.Endpoints["relay"])
		},
	}

	if selectModules == "All" {
		for key, factory := range allModules {
			modules[key] = factory
		}
	} else {
		factory, exists := allModules[selectModules]
		if !exists {
			return nil, fmt.Errorf("selected module '%s' not found", selectModules)
		}
		modules[selectModules] = factory
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	var (
		result = make(map[string]ModulesFasad)
		mu     sync.Mutex
	)

	for name, factory := range modules {
		name := name
		factory := factory

		g.Go(func() error {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			module, err := factory(cfg, clients)
			if err != nil {
				return fmt.Errorf("failed to initialize module %s: %w", name, err)
			}

			mu.Lock()
			result[name] = module
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}
