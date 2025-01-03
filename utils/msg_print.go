package utils

import (
	"lisk/logger"
	"time"
)

func PrintStartMessage() {
	logger.GlobalLogger.Infof("===============================================")
	logger.GlobalLogger.Infof("=    Author Software: @cheifssq               =")
	logger.GlobalLogger.Infof("= Softs, drop checkers and more. Subscribe ;) =")
	logger.GlobalLogger.Infof("===============================================")
	time.Sleep(time.Second * 2)
	logger.GlobalLogger.Warn("Minimum balances to perform activities: USDT/USDC($1) OR WETH($1)")
	time.Sleep(time.Second * 2)
}
