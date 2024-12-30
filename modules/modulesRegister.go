package modules

import (
	"lisk/config"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/modules/dex"
	"lisk/modules/ionic"
	"lisk/modules/relay"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
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
	}

	var (
		result = make(map[string]ModulesFasad)
		errCh  = make(chan error, len(modules))
		mu     sync.Mutex
		wg     sync.WaitGroup
	)

	for name, factory := range modules {
		wg.Add(1)

		go func(name string, factory ModuleFactory) {
			defer wg.Done()

			module, err := factory(cfg, clients)
			if err != nil {
				errCh <- err
				return
			}

			mu.Lock()
			result[name] = module
			mu.Unlock()
		}(name, factory)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return nil, <-errCh
	}

	return result, nil
}
