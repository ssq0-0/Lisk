package utils

import (
	"bufio"
	"fmt"
	"lisk/config"
	"lisk/logger"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func FileReader(filename string) ([]string, error) {
	var lines []string
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func ReadAbis(modulePaths map[string]string) (map[string]*abi.ABI, error) {
	abis := make(map[string]*abi.ABI)

	for module, path := range modulePaths {
		file, err := os.ReadFile(path)
		if err != nil {
			logger.GlobalLogger.Errorf("failed to read abi file: %v, path: %s", err, path)
			return nil, err
		}

		parsedAbi, err := abi.JSON(strings.NewReader(string(file)))
		if err != nil {
			logger.GlobalLogger.Errorf("failed to decode abi: %v, path: %s", err, path)
			return nil, err
		}

		abis[module] = &parsedAbi
	}

	return abis, nil
}

func GetPrivateKeys() ([]string, error) {
	path := getPath("privateKeys")
	keys, err := FileReader(path)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("нет доступных ключей в файле privateKeys.txt")
	}

	return keys, nil
}

func GetConfig() (*config.Config, error) {
	return config.LoadConfig(getPath("config"))
}

func getPath(path string) string {
	paths := map[string]string{
		"privateKeys": "account/privateKeys.txt",
		"config":      "config/config.json",
	}

	return paths[path]
}
