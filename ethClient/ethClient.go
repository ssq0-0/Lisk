package ethClient

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"lisk/account"
	"lisk/globals"
	"lisk/logger"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/sync/errgroup"
)

type Client struct {
	Client *ethclient.Client
}

func EthClientFactory(rpcs map[string]string) (map[string]*Client, error) {
	if len(rpcs) == 0 {
		return nil, errors.New("RPC URLs map is empty")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	var (
		result = make(map[string]*Client)
		mu     sync.Mutex
	)

	for name, rpc := range rpcs {
		name := name
		rpc := rpc

		g.Go(func() error {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			client, err := ethclient.DialContext(ctx, rpc)
			if err != nil {
				return err
			}

			mu.Lock()
			result[name] = &Client{Client: client}
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

func CloseAllClients(clients map[string]*Client) {
	for _, client := range clients {
		if client.Client != nil {
			client.Client.Close()
		}
	}
}

func (c *Client) BalanceCheck(owner, tokenAddr common.Address) (*big.Int, error) {
	if IsNativeToken(tokenAddr) {
		balance, err := c.Client.BalanceAt(context.Background(), owner, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get native coin balance: %v", err)
		}
		return balance, nil
	}

	data, err := globals.Erc20ABI.Pack("balanceOf", owner)
	if err != nil {
		return nil, fmt.Errorf("failed to pack data: %v", err)
	}

	result, err := c.CallCA(tokenAddr, data)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %v", err)
	}

	var balance *big.Int
	if err := globals.Erc20ABI.UnpackIntoInterface(&balance, "balanceOf", result); err != nil {
		return nil, fmt.Errorf("failed to unpack result: %v", err)
	}

	return balance, nil
}

func (c *Client) CallCA(toCA common.Address, data []byte) ([]byte, error) {
	callMsg := ethereum.CallMsg{
		To:   &toCA,
		Data: data,
	}

	return c.Client.CallContract(context.Background(), callMsg, nil)
}

func (c *Client) GetGasValues(msg ethereum.CallMsg) (uint64, *big.Int, *big.Int, error) {
	header, err := c.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, nil, nil, err
	}

	maxPriorityFeePerGas := big.NewInt(1e7)
	maxFeePerGas := new(big.Int).Add(header.BaseFee, maxPriorityFeePerGas)

	gasLimit, err := c.Client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, nil, nil, err
	}

	return gasLimit, maxPriorityFeePerGas, maxFeePerGas, nil
}

func (c *Client) GetNonce(address common.Address) uint64 {
	nonce, err := c.Client.PendingNonceAt(context.Background(), address)
	if err != nil {
		logger.GlobalLogger.Warnf("Failed to get nonce for address %s: %v", address.Hex(), err)
		return 0
	}
	return nonce
}

func (c *Client) GetChainID() (int64, error) {
	chainID, err := c.Client.NetworkID(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get ChainID: %w", err)
	}
	return chainID.Int64(), nil
}

func (c *Client) ApproveTx(tokenAddr, spender common.Address, acc *account.Account, amount *big.Int, rollback bool) (*types.Transaction, error) {
	if IsNativeToken(tokenAddr) {
		return nil, nil
	}

	allowance, err := c.Allowance(tokenAddr, acc.Address, spender)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowance: %v", err)
	}

	var approveValue *big.Int
	if rollback {
		approveValue = big.NewInt(0)
	} else {
		if allowance.Cmp(amount) >= 0 {
			return nil, nil
		}
		approveValue = globals.MaxRepayBigInt
	}

	approveData, err := globals.Erc20ABI.Pack("approve", spender, approveValue)
	if err != nil {
		return nil, fmt.Errorf("failed to pack approve data: %v", err)
	}

	logger.GlobalLogger.Infof("Approve transaction...")
	if err := c.SendTransaction(acc.PrivateKey, acc.Address, tokenAddr, c.GetNonce(acc.Address), big.NewInt(0), approveData); err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 15)
	return nil, nil
}

func (c *Client) Allowance(tokenAddr, owner, spender common.Address) (*big.Int, error) {
	data, err := globals.Erc20ABI.Pack("allowance", owner, spender)
	if err != nil {
		return nil, fmt.Errorf("failed to pack allowance data: %v", err)
	}

	msg := ethereum.CallMsg{
		To:   &tokenAddr,
		Data: data,
	}

	result, err := c.Client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %v", err)
	}

	var allowance *big.Int
	if err = globals.Erc20ABI.UnpackIntoInterface(&allowance, "allowance", result); err != nil {
		return nil, fmt.Errorf("failed to unpack allowance data: %v", err)
	}

	return allowance, nil
}

func (c *Client) SendTransaction(privateKey *ecdsa.PrivateKey, ownerAddr, CA common.Address, nonce uint64, value *big.Int, txData []byte) error {
	chainID, err := c.Client.NetworkID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get ChainID: %v", err)
	}

	gasLimit, maxPriorityFeePerGas, maxFeePerGas, err := c.GetGasValues(ethereum.CallMsg{
		From:  ownerAddr,
		To:    &CA,
		Value: value,
		Data:  txData,
	})
	if err != nil {
		return fmt.Errorf("failed to estimate gas: %v", err)
	}

	dynamicTx := types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: maxPriorityFeePerGas,
		GasFeeCap: maxFeePerGas,
		Gas:       gasLimit,
		To:        &CA,
		Value:     value,
		Data:      txData,
	}

	signedTx, err := types.SignTx(types.NewTx(&dynamicTx), types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	if err = c.Client.SendTransaction(context.Background(), signedTx); err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	logger.GlobalLogger.Infof("[NONCE: %v] Transaction sent: https://blockscout.lisk.com/tx/%s", nonce, signedTx.Hash().Hex())

	return c.waitForTransactionSuccess(signedTx.Hash(), 1*time.Minute)
}

func (c *Client) waitForTransactionSuccess(txHash common.Hash, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.New("transaction wait timeout")
		case <-ticker.C:
			receipt, err := c.Client.TransactionReceipt(context.Background(), txHash)
			if err != nil {
				if err.Error() == "not found" {
					continue
				}
				return fmt.Errorf("error getting transaction receipt: %v", err)
			}

			if receipt.Status == 1 {
				return nil
			} else {
				c.logTransactionError(txHash, receipt)
				return errors.New("transaction failed")
			}
		}
	}
}

func (c *Client) logTransactionError(txHash common.Hash, receipt *types.Receipt) {
	logger.GlobalLogger.Errorf("Transaction failed. txHash: %s", txHash.Hex())

	for _, logEntry := range receipt.Logs {
		logger.GlobalLogger.Warnf("Event Log - Address: %s, Data: %x, Topics: %v",
			logEntry.Address.Hex(),
			logEntry.Data,
			logEntry.Topics,
		)
	}

	tx, isPending, err := c.Client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		logger.GlobalLogger.Warnf("Error getting transaction details: %v", err)
		return
	}

	chainID, err := c.Client.NetworkID(context.Background())
	if err != nil {
		logger.GlobalLogger.Warnf("Error getting ChainID: %v", err)
		return
	}

	from, err := types.Sender(types.LatestSignerForChainID(chainID), tx)
	if err != nil {
		logger.GlobalLogger.Warnf("Error getting transaction sender: %v", err)
		return
	}

	callMsg := ethereum.CallMsg{
		From:     from,
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
		Data:     tx.Data(),
	}

	blockNumber := receipt.BlockNumber
	result, callErr := c.Client.CallContract(context.Background(), callMsg, blockNumber)

	var revertReason string

	if callErr != nil {
		if strings.HasPrefix(callErr.Error(), "execution reverted") {
			revertReason = callErr.Error()
			if len(result) > 0 {
				decodedReason, decodeErr := abi.UnpackRevert(result)
				if decodeErr != nil {
					logger.GlobalLogger.Warnf("Failed to decode revert reason: %v", decodeErr)
				} else {
					revertReason = decodedReason
				}
			}
		} else {
			logger.GlobalLogger.Warnf("Error simulating transaction execution: %v", callErr)
		}
	} else {
		if len(result) > 0 {
			decodedReason, decodeErr := abi.UnpackRevert(result)
			if decodeErr != nil {
				logger.GlobalLogger.Warnf("Failed to decode revert reason: %v", decodeErr)
			} else {
				revertReason = decodedReason
			}
		}
	}

	if revertReason != "" {
		logger.GlobalLogger.Warnf("Transaction revert reason: %s", revertReason)
	} else {
		logger.GlobalLogger.Warnf("Transaction revert reason not found.")
	}

	logger.GlobalLogger.Warnf("Failed transaction details:")
	logger.GlobalLogger.Warnf("  From: %s", from.Hex())
	if tx.To() != nil {
		logger.GlobalLogger.Warnf("  To: %s", tx.To().Hex())
	} else {
		logger.GlobalLogger.Warnf("  To: contract creation (contract transaction)")
	}
	logger.GlobalLogger.Warnf("  Value: %s", tx.Value().String())
	logger.GlobalLogger.Warnf("  Gas Limit: %d", tx.Gas())
	logger.GlobalLogger.Warnf("  Gas Price: %s", tx.GasPrice().String())
	logger.GlobalLogger.Warnf("  Nonce: %d", tx.Nonce())
	logger.GlobalLogger.Warnf("  Data: %x", tx.Data())
	logger.GlobalLogger.Warnf("  Pending: %v", isPending)
}
