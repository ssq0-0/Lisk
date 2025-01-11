package utils

import (
	"fmt"
	"lisk/models"
	"math/big"
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

func ConvertWrapAmount(minStr, maxStr string) (models.WrapRange, error) {
	min, err := ConvertToWei(minStr, 18)
	if err != nil {
		return models.WrapRange{}, err
	}

	max, err := ConvertToWei(maxStr, 18)
	if err != nil {
		return models.WrapRange{}, err
	}

	return models.WrapRange{
		Min: min,
		Max: max,
	}, nil
}
