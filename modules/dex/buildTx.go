package dex

import (
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (d *Dex) buildTxData(tokenIn, tokenOut common.Address, amountIn *big.Int, acc *account.Account, fee *big.Int) ([]byte, [][]byte, error) {
	totalFee, err := d.fetchPool(tokenIn, tokenOut, fee)
	if err != nil {
		return nil, nil, err
	}

	pathBytes, err := d.encodeV3Path(tokenIn, totalFee.Fee, tokenOut)
	if err != nil {
		return nil, nil, err
	}

	amountOutMin, err := d.getAmountMin(pathBytes, amountIn)
	if err != nil {
		return nil, nil, err
	}

	switch {
	case ethClient.IsNativeToken(tokenIn):
		swapData, err := d.packSwapData(acc.Address, amountIn, amountOutMin, pathBytes, false)
		if err != nil {
			return nil, nil, fmt.Errorf("packSwapData failed: %w", err)
		}
		wrapEncoded, err := d.packWrapETHData(d.UniversalCA, amountIn)
		if err != nil {
			return nil, nil, fmt.Errorf("packWrapETHData failed: %w", err)
		}

		commands := []byte{0x0b, 0x00}
		inputs := [][]byte{wrapEncoded, swapData}

		return commands, inputs, nil
	case ethClient.IsNativeToken(tokenOut):
		swapData, err := d.packSwapData(d.UniversalCA, amountIn, amountOutMin, pathBytes, true)
		if err != nil {
			return nil, nil, fmt.Errorf("packSwapData failed: %w", err)
		}
		unwrapEncoded, err := d.packWrapETHData(acc.Address, amountIn)
		if err != nil {
			return nil, nil, fmt.Errorf("packWrapETHData failed: %w", err)
		}
		commands := []byte{0x00, 0x0c}
		inputs := [][]byte{swapData, unwrapEncoded}

		return commands, inputs, nil
	default:
		swapData, err := d.packSwapData(acc.Address, amountIn, amountOutMin, pathBytes, true)
		if err != nil {
			return nil, nil, fmt.Errorf("packSwapData failed: %w", err)
		}
		commands := []byte{0x00}
		inputs := [][]byte{swapData}

		return commands, inputs, nil
	}
}

func (d *Dex) getAmountMin(path []byte, amountIn *big.Int) (*big.Int, error) {
	data, err := d.ABI.Pack("quoteExactInput", path, amountIn)
	if err != nil {
		return nil, fmt.Errorf("failed to pack ABI data: %w", err)
	}

	response, err := d.Client.CallCA(d.Quoter, data)
	if err != nil {
		return nil, fmt.Errorf("call to Quoter failed: %w", err)
	}
	if len(response) == 0 {
		return nil, fmt.Errorf("empty response from contract call")
	}

	unpackedData, err := d.ABI.Unpack("quoteExactInput", response)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack ABI data: %w", err)
	}

	amountMinOut, ok := unpackedData[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("error of conversion to *big.Int")
	}

	return applySlippage(amountMinOut, globals.Slippage), nil
}

func applySlippage(amount *big.Int, slippage *big.Float) *big.Int {
	amountFloat := new(big.Float).SetInt(amount)
	adjustedAmountFloat := new(big.Float).Mul(amountFloat, slippage)
	adjustedAmount, _ := adjustedAmountFloat.Int(nil)
	return adjustedAmount
}
