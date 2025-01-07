package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

func WriteLinesToFile(filePath string, lines []string) error {
	data := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filePath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filePath, err)
	}
	return nil
}

func AppendLinesToFile(filePath string, lines []string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s for appending: %w", filePath, err)
	}
	defer f.Close()

	for _, line := range lines {
		if _, err := f.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to append to file %s: %w", filePath, err)
		}
	}

	return nil
}

func LogErrorAddress(addr, token string, balance string) error {
	errorLine := FormatAddrTokenBalance(addr, token, balance)
	return AppendLinesToFile(GetPath("error"), []string{errorLine})
}

func FormatAddrTokenBalance(addr, token string, balance string) string {
	return fmt.Sprintf("%s,%s,%s", addr, token, balance)
}
