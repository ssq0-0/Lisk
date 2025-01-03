package utils

import (
	"lisk/globals"
	"time"
)

func GetCurrentWeekNumber() int {
	now := time.Now().UTC()
	diffDays := int(now.Sub(globals.StartDate).Hours() / 24)
	return diffDays/7 + 1
}

func GetCurrentYYYYMMDD() int {
	now := time.Now().UTC()
	return now.Year()*10000 + int(now.Month())*100 + now.Day()
}
