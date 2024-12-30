package main

import (
	"fmt"
	"lisk/account"
	"lisk/core/process"
	"lisk/ethClient"
	"lisk/logger"
	"lisk/modules"
	"lisk/utils"
)

func main() {
	privateKeys, err := utils.GetPrivateKeys()
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}

	accs, err := account.AccsFactory(privateKeys)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}

	cfg, err := utils.GetConfig()
	abis, err := utils.ReadAbis(cfg.ABIs)
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

	mods, err := modules.ModulesInit(cfg, abis, clients)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}

	if err := process.ProcessAccount(accs, mods, clients); err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	fmt.Println("Account processed successfully")
}
