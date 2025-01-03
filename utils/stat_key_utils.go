package utils

import (
	"fmt"
	"strconv"
	"strings"

	"lisk/globals"
)

func BuildStatKey(week int, address, module string) globals.StatKey {
	return globals.StatKey(fmt.Sprintf("%d|%s|%s", week, address, module))
}

func ParseStatKey(key globals.StatKey) (int, string, string, error) {
	parts := strings.Split(string(key), "|")
	if len(parts) != 3 {
		return 0, "", "", fmt.Errorf("invalid key format: %s", key)
	}

	week, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to parse week: %w", err)
	}

	return week, parts[1], parts[2], nil
}
