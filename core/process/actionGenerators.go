package process

import (
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/logger"
	"lisk/models"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var actionGenerators = map[string]func(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error){
	"Oku":                generateSwap,
	"Ionic":              generateLiquidity,
	"IonicWithdraw":      generateIonicWithdraw,
	"Relay":              generateBridgeToLisk,
	"Checker":            generateChecker,
	"Portal_daily_check": generateDailyCheck,
	"Portal_main_tasks":  generateMainTasks,
}

func generateChecker(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	return ActionProcess{TypeAction: globals.Checker, Module: "Portal"}, nil
}
func generateDailyCheck(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	return ActionProcess{TypeAction: globals.DailyCheck, Module: "Portal"}, nil
}

func generateMainTasks(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	return ActionProcess{TypeAction: globals.MainTasks, Module: "Portal"}, nil
}

func generateSwap(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	if len(acc.LastSwaps) > 1 {
		last := acc.LastSwaps[len(acc.LastSwaps)-1]
		prev := acc.LastSwaps[len(acc.LastSwaps)-2]
		if last.TokenFrom == prev.TokenFrom && last.TokenTo == prev.TokenTo {
			return ActionProcess{TypeAction: globals.Unknown},
				fmt.Errorf("two identical swaps in a row not allowed")
		}
	}

	var tokenFrom common.Address
	if len(acc.LastSwaps) == 0 {
		tokenFrom = globals.WETH
	} else {
		lastTokenTo := acc.LastSwaps[len(acc.LastSwaps)-1].TokenTo
		if lastTokenTo != globals.LISK {
			tokenFrom = lastTokenTo
		} else {
			tokenFrom = globals.WETH
		}
	}

	var tokenTo common.Address
	const maxAttempts = 5
	for attempts := 0; attempts < maxAttempts; attempts++ {
		tokenTo = selectDifferentToken(tokenFrom)
		if !excludeSwap(tokenFrom, tokenTo) {
			break
		}
		if attempts == maxAttempts-1 {
			return ActionProcess{TypeAction: globals.Unknown}, fmt.Errorf("failed to select non-excluded token after %d attempts", maxAttempts)
		}
	}
	acc.LastSwaps = append(acc.LastSwaps, models.SwapPair{
		TokenFrom: tokenFrom,
		TokenTo:   tokenTo,
	})

	amount, err := canDoActionByBalance(tokenFrom, acc, clients["lisk"])
	if err != nil {
		return ActionProcess{TypeAction: globals.Unknown}, err
	}

	return ActionProcess{
		TokenFrom:  tokenFrom,
		TokenTo:    tokenTo,
		Amount:     amount,
		TypeAction: globals.Swap,
		Module:     "Oku",
	}, nil
}

func generateLiquidity(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	liqState := acc.LiquidityState

	if liqState.ActionCount == 0 {
		action := ActionProcess{
			TokenFrom:  globals.USDT,
			Amount:     big.NewInt(1e6),
			TypeAction: globals.Supply,
			Module:     "Ionic",
		}
		updateLiquidityState(acc, action.TypeAction)
		return action, nil
	}

	if liqState.ActionCount == 1 {
		action := ActionProcess{
			TypeAction: globals.EnterMarket,
			Module:     "Ionic",
		}
		updateLiquidityState(acc, action.TypeAction)
		return action, nil
	}

	if liqState.LastAction == globals.Redeem {
		if !liqState.PendingEnterAfterWithdraw {
			liqState.PendingEnterAfterWithdraw = true
			action := ActionProcess{
				TokenFrom:  globals.USDT,
				Amount:     big.NewInt(1e5),
				TypeAction: globals.Supply,
				Module:     "Ionic",
			}
			updateLiquidityState(acc, action.TypeAction)
			return action, nil
		}
	}

	if liqState.LastAction == globals.Borrow {
		action := ActionProcess{
			TokenFrom:  globals.LISK,
			Amount:     big.NewInt(19e16),
			TypeAction: globals.Repay,
			Module:     "Ionic",
		}
		updateLiquidityState(acc, action.TypeAction)
		return action, nil
	}

	if liqState.LastAction == globals.Repay || liqState.LastAction == globals.EnterMarket {
		action := ActionProcess{
			TokenFrom:  globals.LISK,
			Amount:     big.NewInt(2e17),
			TypeAction: globals.Borrow,
			Module:     "Ionic",
		}
		updateLiquidityState(acc, action.TypeAction)
		return action, nil
	}

	return ActionProcess{TypeAction: globals.Unknown}, nil
}

func generateIonicWithdraw(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	if acc.LiquidityState.ActionCount == 0 {
		actions := ActionProcess{
			TokenFrom:  globals.USDT,
			TypeAction: globals.ExitMarket,
			Module:     "Ionic",
		}

		updateLiquidityState(acc, globals.ExitMarket)
		return actions, nil
	}

	if acc.LiquidityState.LastAction == globals.ExitMarket {
		actions := ActionProcess{
			Amount:     globals.MaxRepayBigInt,
			TokenFrom:  globals.USDT,
			TypeAction: globals.Redeem,
			Module:     "Ionic",
		}

		updateLiquidityState(acc, globals.Redeem)
		return actions, nil
	}

	return ActionProcess{}, fmt.Errorf("Unknow action")
}

func generateBridgeToLisk(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	chain, balance, err := getMaxBalance(acc, clients)
	if err != nil {
		return ActionProcess{TypeAction: globals.Unknown}, fmt.Errorf("failed get max balance in all chains: %v", err)
	}

	if chain == "" {
		return ActionProcess{TypeAction: globals.Unknown}, fmt.Errorf("no balance in other networks")
	}
	logger.GlobalLogger.Infof("Bridge to LISK. From: %s.", chain)
	return bridgeToLisk(acc, balance, chain, clients[chain])
}

func bridgeToLisk(acc *account.Account, balance *big.Int, chain string, client *ethClient.Client) (ActionProcess, error) {
	percentBalance := new(big.Int).Mul(balance, big.NewInt(70))
	percentBalance.Div(percentBalance, big.NewInt(100))

	return ActionProcess{
		TokenFrom:  globals.NATIVE,
		TypeAction: globals.Bridge[chain],
		Amount:     percentBalance,
		Module:     "Relay",
	}, nil
}
