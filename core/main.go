package main

import (
	"fmt"
	"lisk/account"
	"lisk/core/process"
	"lisk/ethClient"
	"lisk/logger"
	"lisk/modules"
	"lisk/utils"
	"regexp"
	"time"

	"github.com/AlecAivazis/survey/v2"
)

func main() {
	utils.PrintStartMessage()

	if err := utils.CheckVersion(); err != nil {
		logger.GlobalLogger.Warn(err)
	}

	privateKeys, err := utils.GetPrivateKeys()
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	logger.GlobalLogger.Infof("Private keys have been initialised!")

	cfg, err := utils.GetConfig()
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	logger.GlobalLogger.Infof("Config have been initialised!")

	clients, err := ethClient.EthClientFactory(cfg.RPC)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	defer ethClient.CloseAllClients(clients)
	logger.GlobalLogger.Infof("Ethereum clients have been initialised!")

	abis, err := utils.ReadAbis(cfg.ABIs)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	logger.GlobalLogger.Infof("ABI's have been initialised!")

	selectModule := userChoice()
	if selectModule == "" {
		logger.GlobalLogger.Infof("No module selected or an error occurred. Exiting.")
		return
	}

	mods, err := modules.ModulesInit(cfg, abis, clients)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	logger.GlobalLogger.Infof("Modules have been initialised!")

	if selectModule == "Exit" {
		logger.GlobalLogger.Infof("Finishing the programme.")
		return
	}
	logger.GlobalLogger.Infof("The %s module was chosen", selectModule)

	proxys, err := utils.GetProxys()
	if err != nil {
		logger.GlobalLogger.Warn(err)
	}

	accs, err := account.AccsFactory(privateKeys, proxys, cfg)
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

func userChoice() string {
	modules := []string{
		"1. Oku",
		"2. Ionic",
		"3. IonicWithdraw",
		"4. Relay",
		"5. Checker",
		"6. Portal_daily_check",
		"7. Portal_main_tasks",
		"0. Exit",
	}

	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: "Choose module",
		Options: modules,
		Default: modules[len(modules)-1],
	}, &selected); err != nil {
		logger.GlobalLogger.Errorf("Ошибка выбора модуля: %v", err)
		return ""
	}

	rgx := regexp.MustCompile(`^\d+\.\s*`)
	selected = rgx.ReplaceAllString(selected, "")
	return selected
}
