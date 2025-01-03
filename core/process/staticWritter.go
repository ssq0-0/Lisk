package process

import (
	"fmt"
	"lisk/account"
	"lisk/models"
	"lisk/utils"
)

func WriteWeeklyStats(accounts []*account.Account) error {
	csvFilePath := utils.GetPath("stats")

	statsMap, err := utils.ReadStatsFromCSV(csvFilePath)
	if err != nil {
		return fmt.Errorf("failed to read stats from CSV: %w", err)
	}

	currentWeek := utils.GetCurrentWeekNumber()
	thisDate := utils.GetCurrentYYYYMMDD()

	for _, acc := range accounts {
		address := acc.Address.Hex()

		for moduleName, delta := range acc.Stats {
			if delta < 0 {
				delta = 0
			}

			key := utils.BuildStatKey(currentWeek, address, moduleName)
			oldRecord, exists := statsMap[key]

			if !exists {
				oldRecord = models.StatRecord{
					TotalSuccess: 0,
					TodaySuccess: 0,
					TodayDate:    thisDate,
				}
			}

			if oldRecord.TodayDate == thisDate {
				oldRecord.TodaySuccess += delta
			} else {
				oldRecord.TodaySuccess = delta
				oldRecord.TodayDate = thisDate
			}

			oldRecord.TotalSuccess += delta

			statsMap[key] = oldRecord
		}
	}

	if err := utils.WriteStatsToCSV(csvFilePath, statsMap); err != nil {
		return fmt.Errorf("failed to write stats to CSV: %w", err)
	}

	return nil
}
