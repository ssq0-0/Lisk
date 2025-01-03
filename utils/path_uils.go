package utils

func GetPath(path string) string {
	paths := map[string]string{
		"privateKeys": "account/privateKeys.txt",
		"config":      "config/config.json",
		"proxy":       "account/proxy.txt",
		"stats":       "account/account_stats.csv",
	}

	return paths[path]
}
