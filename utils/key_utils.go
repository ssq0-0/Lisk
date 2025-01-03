package utils

import (
	"fmt"
)

func GetPrivateKeys() ([]string, error) {
	path := GetPath("privateKeys")
	keys, err := FileReader(path)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("нет доступных ключей в файле privateKeys.txt")
	}

	return keys, nil
}

func GetProxies() ([]string, error) {
	path := GetPath("proxy")
	proxies, err := FileReader(path)
	if err != nil {
		return nil, err
	}

	if len(proxies) == 0 {
		return nil, fmt.Errorf("нет доступных прокси в файле proxy.txt")
	}

	return proxies, nil
}
