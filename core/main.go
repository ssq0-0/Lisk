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

	if selectModule == "Exit" {
		logger.GlobalLogger.Infof("Finishing the programme.")
		return
	}
	logger.GlobalLogger.Infof("The %s module was chosen", selectModule)

	mods, err := modules.ModulesInit(cfg, selectModule, abis, clients)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	logger.GlobalLogger.Infof("Modules have been initialised!")

	accs, err := account.AccsFactory(privateKeys, cfg)
	if err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	logger.GlobalLogger.Infof("All settings are initialised! Sleep 5 seconds...")
	time.Sleep(time.Second * 5)

	if err := process.ProcessAccounts(accs, mods, clients); err != nil {
		logger.GlobalLogger.Error(err)
		return
	}
	fmt.Println("Account processed successfully")
}

func userChoice() string {
	modules := []string{
		"1. Oku",
		"2. Ionic",
		"3. Relay",
		"4. All",
		"5. Checker",
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
