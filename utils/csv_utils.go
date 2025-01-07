package utils

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"lisk/globals"
	"lisk/models"
)

func ReadStatsFromCSV(filePath string) (map[globals.StatKey]models.StatRecord, error) {
	statsMap := make(map[globals.StatKey]models.StatRecord)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return statsMap, nil
		}
		return nil, fmt.Errorf("failed to open CSV file for reading: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil {
		if err != io.EOF {
			return nil, fmt.Errorf("failed to read CSV header: %w", err)
		}
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV record: %w", err)
		}

		if len(record) != 6 {
			continue
		}

		week, err := strconv.Atoi(strings.TrimSpace(record[0]))
		if err != nil {
			continue
		}
		address := strings.TrimSpace(record[1])
		module := strings.TrimSpace(record[2])

		totalSuccess, err := strconv.Atoi(strings.TrimSpace(record[3]))
		if err != nil {
			continue
		}
		todaySuccess, err := strconv.Atoi(strings.TrimSpace(record[4]))
		if err != nil {
			continue
		}
		todayDate, err := strconv.Atoi(strings.TrimSpace(record[5]))
		if err != nil {
			continue
		}

		key := BuildStatKey(week, address, module)
		statsMap[key] = models.StatRecord{
			TotalSuccess: totalSuccess,
			TodaySuccess: todaySuccess,
			TodayDate:    todayDate,
		}
	}

	return statsMap, nil
}

func WriteStatsToCSV(filePath string, statsMap map[globals.StatKey]models.StatRecord) error {
	tempFilePath := filePath + ".tmp"
	file, err := os.Create(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to create temporary CSV file: %w", err)
	}
	defer func() {
		file.Close()
		os.Remove(tempFilePath)
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Week", "AccountAddress", "Module", "TotalSuccess", "TodaySuccess", "TodayDate"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	keys := make([]globals.StatKey, 0, len(statsMap))
	for key := range statsMap {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		_, addr1, mod1, err1 := ParseStatKey(keys[i])
		_, addr2, mod2, err2 := ParseStatKey(keys[j])

		if err1 != nil || err2 != nil {
			return false
		}

		if addr1 == addr2 {
			return mod1 < mod2
		}
		return addr1 < addr2
	})

	for _, key := range keys {
		stats := statsMap[key]
		week, address, module, err := ParseStatKey(key)
		if err != nil {
			continue
		}

		record := []string{
			strconv.Itoa(week),
			address,
			module,
			strconv.Itoa(stats.TotalSuccess),
			strconv.Itoa(stats.TodaySuccess),
			strconv.Itoa(stats.TodayDate),
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing CSV writer: %w", err)
	}
	file.Close()

	if err := os.Rename(tempFilePath, filePath); err != nil {
		return fmt.Errorf("failed to rename temporary CSV file: %w", err)
	}

	return nil
}

func ReplacePrivateKey(pk, addr string) error {
	lines, err := FileReader(GetPath("privateKeys"))
	if err != nil {
		return fmt.Errorf("failed to read private keys file: %w", err)
	}

	updatedLines := make([]string, 0, len(lines))
	keyFound := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == pk {
			keyFound = true
			continue
		}

		updatedLines = append(updatedLines, line)
	}

	if !keyFound {
		return fmt.Errorf("private key not found in the file")
	}

	if err := WriteLinesToFile(GetPath("privateKeys"), updatedLines); err != nil {
		return fmt.Errorf("failed to update private keys file: %w", err)
	}

	errorLine := fmt.Sprintf("%s,%s", addr, pk)
	if err := AppendLinesToFile(GetPath("error"), []string{errorLine}); err != nil {
		return fmt.Errorf("failed to append error address: %w", err)
	}

	return nil
}

func WriteAddrTokenBalancesToFile(balances map[string]map[string]string) error {
	lines := make([]string, 0)

	for addr, tokens := range balances {
		for token, balance := range tokens {
			line := FormatAddrTokenBalance(addr, token, balance)
			lines = append(lines, line)
		}
	}

	return AppendLinesToFile(GetPath("balances"), lines)
}
