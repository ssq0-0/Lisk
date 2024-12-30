package dex

import (
	"fmt"
	"lisk/account"
	"lisk/globals"
	"lisk/logger"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func (d *Dex) ensureAllowance(token common.Address, amount *big.Int, acc *account.Account) error {
	if err := d.ensurePermitAllowance(token, amount, acc); err != nil {
		return fmt.Errorf("permit allowance check failed: %w", err)
	}

	if err := d.ensureRouterAllowance(token, amount, acc); err != nil {
		return fmt.Errorf("router allowance check failed: %w", err)
	}

	return nil
}

func (d *Dex) ensurePermitAllowance(token common.Address, amount *big.Int, acc *account.Account) error {
	data, err := globals.Erc20ABI.Pack("allowance", acc.Address, d.PermitCA)
	if err != nil {
		return fmt.Errorf("failed to pack permit allowance data: %w", err)
	}

	result, err := d.Client.CallCA(token, data)
	if err != nil {
		return fmt.Errorf("permit allowance call failed: %w", err)
	}

	unpackedData, err := globals.Erc20ABI.Methods["allowance"].Outputs.Unpack(result)
	if err != nil {
		return fmt.Errorf("failed to unpack permit allowance data: %w", err)
	}

	permitAllowance, ok := unpackedData[0].(*big.Int)
	if !ok {
		return fmt.Errorf("unexpected type for permit allowance")
	}

	if permitAllowance.Cmp(amount) < 0 {
		if _, err := d.Client.ApproveTx(token, d.PermitCA, acc, globals.MaxApprove, false); err != nil {
			return fmt.Errorf("failed to approve via permit: %w", err)
		}
	}

	return nil
}

func (d *Dex) ensureRouterAllowance(token common.Address, amount *big.Int, acc *account.Account) error {
	data, err := d.ABI.Pack("allowance", acc.Address, token, d.UniversalCA)
	if err != nil {
		return fmt.Errorf("failed to pack router allowance data: %w", err)
	}

	result, err := d.Client.CallCA(d.PermitCA, data)
	if err != nil {
		return fmt.Errorf("router allowance call failed: %w", err)
	}

	unpackedData, err := d.ABI.Methods["allowance"].Outputs.Unpack(result)
	if err != nil {
		return fmt.Errorf("failed to unpack router allowance data: %w", err)
	}

	if len(unpackedData) < 2 {
		return fmt.Errorf("unexpected result: insufficient data for router allowance")
	}

	routerAllowance, ok := unpackedData[0].(*big.Int)
	if !ok {
		return fmt.Errorf("unexpected type for router allowance")
	}

	expiration, ok := unpackedData[1].(*big.Int)
	if !ok {
		return fmt.Errorf("unexpected type for expiration")
	}

	currentTime := big.NewInt(time.Now().Unix())
	if routerAllowance.Cmp(amount) < 0 || expiration.Cmp(currentTime) <= 0 {
		logger.GlobalLogger.Infof("Router allowance insufficient or expired. Initiating approval...")
		if err := d.approveToken(token, acc); err != nil {
			return fmt.Errorf("failed to approve router allowance: %w", err)
		}
		time.Sleep(20 * time.Second)
	}

	return nil
}

func (d *Dex) approveToken(token common.Address, acc *account.Account) error {
	deadline := big.NewInt(time.Now().Unix() + int64(globals.ApproveDeadlineOffset))
	data, err := d.ABI.Pack("approve", token, d.UniversalCA, globals.MaxApprove, deadline)
	if err != nil {
		return fmt.Errorf("failed to pack approve data: %w", err)
	}
	return d.Client.SendTransaction(acc.PrivateKey, acc.Address, d.PermitCA, d.Client.GetNonce(acc.Address), big.NewInt(0), data)
}
