package utils

import (
	"lisk/config"
)

func GetConfig() (*config.Config, error) {
	return config.LoadConfig(GetPath("config"))
}
