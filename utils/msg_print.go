package utils

import (
	"lisk/globals"
	"lisk/httpClient"
	"lisk/logger"
	"lisk/models"
	"time"
)

func PrintStartMessage() {
	logger.GlobalLogger.Infof("===============================================")
	logger.GlobalLogger.Infof("=    Author Software: @cheifssq               =")
	logger.GlobalLogger.Infof("= Softs, drop checkers and more. Subscribe ;) =")
	logger.GlobalLogger.Infof("===============================================")
	time.Sleep(time.Second * 2)
}

func GasPricesPrint() {
	client, err := httpClient.NewHttpClient("")
	if err != nil {
		return
	}

	var result models.BlockscoutResp
	if err := client.SendJSONRequest(globals.Blockscout, "GET", nil, &result); err != nil {
		return
	}
	logger.GlobalLogger.Infof("==========================================================")
	logger.GlobalLogger.Infof("=Current gwei in LISK| SLOW %v| AVERAGE %v| FAST %v=", result.GasPrice.Slow, result.GasPrice.Average, result.GasPrice.Fast)
	logger.GlobalLogger.Infof("==========================================================")
	time.Sleep(time.Second * 3)
}
