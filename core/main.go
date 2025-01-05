package main

import (
	"fmt"
	"lisk/account"
	"lisk/core/process"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/logger"
	"lisk/modules"
	"lisk/utils"
	"time"
)

func main() {
	_ = utils.SetConsoleTitle(globals.ConsoleTitle)

	utils.PrintStartMessage()
	utils.GasPricesPrint()
	if err := utils.CheckVersion(); err != nil {
		logger.GlobalLogger.Warn(err)
	}

	privateKeys, err := utils.GetPrivateKeys()
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}

	cfg, err := utils.GetConfig()
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}

	process.InitGlobals(cfg)

	clients, err := ethClient.EthClientFactory(cfg.RPC)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	defer ethClient.CloseAllClients(clients)

	abis, err := utils.ReadAbis(cfg.ABIs)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}

	memory, err := process.NewMemory(cfg.StateFile)
	if err != nil {
		logger.GlobalLogger.Warn(err)
		return
	}

	selectModule, err := determineModuleForRun(memory)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}

	mods, err := modules.ModulesInit(cfg, abis, clients)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}

	proxies, err := utils.GetProxies()
	if err != nil {
		logger.GlobalLogger.Warn(err)
	}

	accs, err := account.AccsFactory(privateKeys, proxies, cfg)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	logger.GlobalLogger.Infof("All settings are initialised! Sleep 5 seconds...")
	time.Sleep(time.Second * 5)

	if err := process.ProcessAccounts(accs, selectModule, mods, clients, memory); err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	logger.GlobalLogger.Infof("Account processed successfully")
	utils.PrintStartMessage()
}

func determineModuleForRun(memory *process.Memory) (string, error) {
	hasSavedState, err := memory.IsStateFileNotEmpty()
	if err != nil {
		return "", fmt.Errorf("failed to check state file: %w", err)
	}

	if hasSavedState {
		answr := utils.ResotoreProcess()
		if answr == "Yes" {
			return "", nil
		} else {
			if err := memory.ClearAllStates(); err != nil {
				return "", fmt.Errorf("failed to clear state file: %w", err)
			}

			selectModule := utils.UserChoice()
			if selectModule == "" || selectModule == "Exit" {
				return "", fmt.Errorf("No module selected. Exiting.")
			}
			return selectModule, nil
		}
	}

	selectModule := utils.UserChoice()
	if selectModule == "" || selectModule == "Exit" {
		return "", fmt.Errorf("No module selected. Exiting.")
	}
	return selectModule, nil
}
