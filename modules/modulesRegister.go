package modules

import (
	"context"
	"fmt"
	"lisk/config"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/httpClient"
	"lisk/modules/balanceChecker"
	"lisk/modules/dex"
	"lisk/modules/eligbleChecker"
	"lisk/modules/ionic"
	"lisk/modules/liskPortal"
	"lisk/modules/relay"
	"lisk/modules/wraper"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"
)

type ModuleFactory func(cfg *config.Config, clients map[string]*ethClient.Client) (ModulesFasad, error)

func ModulesInit(cfg *config.Config, abis map[string]*abi.ABI, clients map[string]*ethClient.Client) (map[string]ModulesFasad, error) {
	modules := map[string]ModuleFactory{
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
		"Portal": func(cfg *config.Config, clients map[string]*ethClient.Client) (ModulesFasad, error) {
			return liskPortal.NewPortal(cfg.Endpoints["lisk_portal"], cfg.Endpoints["top"])
		},
		"Wraper": func(cfg *config.Config, clients map[string]*ethClient.Client) (ModulesFasad, error) {
			return wraper.NewWraper(clients["lisk"])
		},
		"Balances": func(cfg *config.Config, clients map[string]*ethClient.Client) (ModulesFasad, error) {
			tokens := map[string]common.Address{
				"ETH":  globals.WETH,
				"USDT": globals.USDT,
				"USDC": globals.USDC,
				"LISK": globals.LISK,
			}
			return balanceChecker.NewChecker(clients["lisk"], tokens)
		},
		"AirdropStatus": func(cfg *config.Config, clients map[string]*ethClient.Client) (ModulesFasad, error) {
			hc, err := httpClient.NewHttpClient("")
			if err != nil {
				return nil, err
			}

			return eligbleChecker.NewChecker(hc)
		},
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
