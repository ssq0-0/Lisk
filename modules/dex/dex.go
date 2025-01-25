package dex

import (
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Dex struct {
	ABI         *abi.ABI
	UniversalCA common.Address // Адрес Universal Router
	PermitCA    common.Address
	Factory     common.Address
	Quoter      common.Address
	Fees        []*big.Int // например 3000 (0.3%)
	Client      *ethClient.Client
}

func NewDex(addresses map[string]string, univAbi *abi.ABI, client *ethClient.Client) (*Dex, error) {
	universalCA := common.HexToAddress(addresses["swap_router"])
	if universalCA == (common.Address{}) {
		return nil, fmt.Errorf("invalid 'swap_router' address")
	}
	permitCA := common.HexToAddress(addresses["permit"])
	if permitCA == (common.Address{}) {
		return nil, fmt.Errorf("invalid 'permit' address")
	}
	factoryCA := common.HexToAddress(addresses["factory"])
	if factoryCA == (common.Address{}) {
		return nil, fmt.Errorf("invalid 'factory' address")
	}
	quoterCA := common.HexToAddress(addresses["quoter"])
	if quoterCA == (common.Address{}) {
		return nil, fmt.Errorf("invalid 'quoter' address")
	}

	fees := []*big.Int{
		big.NewInt(100),
		big.NewInt(200),
		big.NewInt(300),
		big.NewInt(500),
		big.NewInt(1000),
		big.NewInt(3000),
	}

	return &Dex{
		ABI:         univAbi,
		UniversalCA: universalCA,
		PermitCA:    permitCA,
		Factory:     factoryCA,
		Quoter:      quoterCA,
		Fees:        fees,
		Client:      client,
	}, nil
}

func (d *Dex) Action(tokenIn, tokenOut common.Address, amountIn *big.Int, acc *account.Account, actionType globals.ActionType) error {
	data, value, err := d.createTransaction(tokenIn, tokenOut, amountIn, acc)
	if err != nil {
		return err
	}

	if !ethClient.IsNativeToken(tokenIn) {
		if err := d.ensureAllowance(tokenIn, amountIn, acc); err != nil {
			return fmt.Errorf("failed to approve tokens: %w", err)
		}
	}

	return d.Client.SendTransaction(
		acc.PrivateKey,
		acc.Address,
		d.UniversalCA,
		d.Client.GetNonce(acc.Address),
		value,
		data,
	)
}

func (d *Dex) createTransaction(tokenIn, tokenOut common.Address, amountIn *big.Int, acc *account.Account) ([]byte, *big.Int, error) {
	var (
		localFees = append([]*big.Int{}, d.Fees...)
		commands  []byte
		inputs    [][]byte
		err       error
	)

	for i := 0; i < len(localFees); i++ {
		fee := localFees[i]
		commands, inputs, err = d.buildTxData(tokenIn, tokenOut, amountIn, acc, fee)
		if err != nil {
			if verifyError(err) {
				localFees = removeFee(localFees, i)
				i--
				continue
			}
			return nil, nil, fmt.Errorf("error building transaction data: %w", err)
		}
	}
	deadline := big.NewInt(time.Now().Unix() + int64(globals.DefaultDeadlineOffset)) // 20 min
	data, err := d.ABI.Pack("execute", commands, inputs, deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to pack universalRouter.execute: %w", err)
	}

	value := big.NewInt(0)
	if ethClient.IsNativeToken(tokenIn) {
		value = amountIn
	}

	return data, value, nil
}

func removeFee(fees []*big.Int, index int) []*big.Int {
	return append(fees[:index], fees[index+1:]...)
}

func verifyError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	errorSubstrings := []string{
		"no pool found for tokens",
		"invalid price calculated",
		"invalid pool address returned",
		"call to Quoter failed: execution reverted",
		// "Gas wait timeout has been exceeded",
	}

	for _, substr := range errorSubstrings {
		if strings.Contains(errMsg, substr) {
			return true
		}
	}
	return false
}
