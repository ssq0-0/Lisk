package process

import (
	"lisk/config"
	"lisk/globals"
	"lisk/logger"
	"lisk/utils"
	"math/big"
	"time"
)

func InitGlobals(cfg *config.Config) {
	if cfg.StartDate != "" {
		if parsedDate, err := parseDate(cfg.StartDate); err == nil {
			globals.StartDate = parsedDate
		} else {
			logger.GlobalLogger.Errorf("failed parse StartDate: %v", err)
		}
	}

	initGlobalWei(&globals.AttentionGwei, cfg.AttentionGwei, 9, "AttantionGwei")
	initGlobalWei(&globals.IonicBorrow, cfg.IonicBorrow, 18, "IonicBorrow")
	initGlobalWei(&globals.IonicSupply, cfg.IonicSupply, 6, "IonicSupply")

	initGlobalDuration(&globals.AttentionTime, cfg.AttentionTime, "AttantionTime")
	initGlobalDuration(&globals.MaxAttentionTime, cfg.MaxAttentionTime, "MaxAttantionTime")
}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, dateStr)
}

func initGlobalWei(globalVar **big.Int, value string, decimals int, name string) {
	if value != "" {
		convertedValue, err := utils.ConvertToWei(value, decimals)
		if err != nil {
			logger.GlobalLogger.Errorf("failed convert type %s: %v", name, err)
			return
		}
		*globalVar = convertedValue
	}
}

func initGlobalDuration(globalVar *int, value int, name string) {
	if value != 0 {
		*globalVar = value
	}
}
