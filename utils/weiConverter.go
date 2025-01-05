package utils

import (
	"fmt"
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
