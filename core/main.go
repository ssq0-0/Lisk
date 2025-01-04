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

	clients, err := ethClient.EthClientFactory(cfg.RPC)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	defer ethClient.CloseAllClients(clients)

	clients["lisk"].PrintGasInStart()

	abis, err := utils.ReadAbis(cfg.ABIs)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}

	selectModule := utils.UserChoice()
	if selectModule == "" || selectModule == "Exit" {
		logger.GlobalLogger.Infof("No module selected or an error occurred. Exiting.")
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

	if err := process.ProcessAccounts(accs, selectModule, mods, clients); err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	fmt.Println("Account processed successfully")
}
