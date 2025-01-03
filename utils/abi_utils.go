package utils

import (
	"fmt"
	"os"
	"strings"

	"lisk/logger"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func ReadAbis(modulePaths map[string]string) (map[string]*abi.ABI, error) {
	abis := make(map[string]*abi.ABI)

	for module, path := range modulePaths {
		data, err := os.ReadFile(path)
		if err != nil {
			logger.GlobalLogger.Errorf("failed to read ABI file: %v, path: %s", err, path)
			return nil, err
		}

		parsedAbi, err := abi.JSON(strings.NewReader(string(data)))
		if err != nil {
			logger.GlobalLogger.Errorf("failed to parse ABI: %v, path: %s", err, path)
			return nil, err
		}

		abis[module] = &parsedAbi
	}

	if len(abis) == 0 {
		return nil, fmt.Errorf("no ABIs were loaded. Check module paths")
	}

	return abis, nil
}
