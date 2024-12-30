package dex

import (
	"fmt"
	"lisk/globals"
	"lisk/models"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (d *Dex) getFee(tokenIn, tokenOut common.Address) (*big.Int, error) {
	for _, fee := range d.Fees {
		pool, err := d.fetchPool(tokenIn, tokenOut, fee)
		if err != nil {
			continue
		}
		if pool.PoolAddress != (common.Address{}) {
			return fee, nil
		}
	}

	return nil, fmt.Errorf("no pool found for tokens %s and %s", tokenIn.Hex(), tokenOut.Hex())
}

func (d *Dex) fetchPool(tokenIn, tokenOut common.Address, fee *big.Int) (*models.FeePool, error) {
	data, err := d.ABI.Pack("getPool", tokenIn, tokenOut, fee)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getPool data: %w", err)
	}

	result, err := d.Client.CallCA(d.Factory, data)
	if err != nil {
		return nil, fmt.Errorf("getPool call failed: %w", err)
	}

	unpacked, err := d.ABI.Methods["getPool"].Outputs.Unpack(result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getPool result: %w", err)
	}

	poolAddress, ok := unpacked[0].(common.Address)
	if !ok || poolAddress == (common.Address{}) {
		return nil, fmt.Errorf("invalid pool address returned")
	}

	data, err = d.ABI.Pack("slot0")
	if err != nil {
		return nil, fmt.Errorf("failed to pack slot0 data: %w", err)
	}

	result, err = d.Client.CallCA(poolAddress, data)
	if err != nil {
		return nil, fmt.Errorf("slot0 call failed: %w", err)
	}

	unpackedSlot0, err := d.ABI.Methods["slot0"].Outputs.Unpack(result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack slot0 result: %w", err)
	}

	sqrtPriceX96, ok := unpackedSlot0[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid sqrtPriceX96 type")
	}

	price := calculatePrice(
		sqrtPriceX96,
		globals.TokensDecimals[poolAddress].Token0,
		globals.TokensDecimals[poolAddress].Token1,
	)
	if price <= 0 {
		return nil, fmt.Errorf("invalid price calculated: %f", price)
	}

	return &models.FeePool{
		Fee:          fee,
		PoolAddress:  poolAddress,
		SqrtPriceX96: sqrtPriceX96,
	}, nil
}

func calculatePrice(sqrtPriceX96 *big.Int, token0Decimals, token1Decimals int) float64 {
	sqrtPrice := new(big.Float).SetInt(sqrtPriceX96)
	scaleFactor := new(big.Float).SetFloat64(math.Pow(2, 96))
	price := new(big.Float).Quo(sqrtPrice, scaleFactor)
	priceSquared := new(big.Float).Mul(price, price)

	decimalAdjustment := float64(token1Decimals - token0Decimals)
	decimalFactor := math.Pow(10, decimalAdjustment)
	priceSquared.Mul(priceSquared, big.NewFloat(decimalFactor))

	finalPrice, _ := priceSquared.Float64()

	return finalPrice
}
