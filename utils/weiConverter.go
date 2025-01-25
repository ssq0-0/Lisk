package utils

import (
	"fmt"
	"lisk/models"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func ConvertToWei(amount string, decimals int) (*big.Int, error) {
	amountFloat, _, err := big.ParseFloat(amount, 10, 0, big.ToNearestEven)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %v", err)
	}

	multiplier := new(big.Float).SetFloat64(float64(1))
	for i := 0; i < decimals; i++ {
		multiplier.Mul(multiplier, new(big.Float).SetFloat64(10))
	}

	amountWei := new(big.Float).Mul(amountFloat, multiplier)

	wei := new(big.Int)
	amountWei.Int(wei)

	return wei, nil
}

func ConvertFromWei(wei *big.Int, decimals int) string {
	weiFloat := new(big.Float).SetInt(wei)

	divisor := new(big.Float).SetFloat64(1)
	for i := 0; i < decimals; i++ {
		divisor.Mul(divisor, new(big.Float).SetFloat64(10))
	}

	result := new(big.Float).Quo(weiFloat, divisor)

	return result.Text('f', decimals)
}

func ConvertRangeAmount(minStr, maxStr string, decimals int, token common.Address, model interface{}) error {
	min, err := ConvertToWei(minStr, decimals)
	if err != nil {
		return err
	}

	max, err := ConvertToWei(maxStr, decimals)
	if err != nil {
		return err
	}

	switch m := model.(type) {
	case *models.WrapRange:
		m.Min = min
		m.Max = max
	case *models.SwapRange:
		m.MinSwapAmount[token] = min
		m.MaxSwapAmount[token] = max
	default:
		return fmt.Errorf("неподдерживаемый тип модели")
	}
	return nil
}
