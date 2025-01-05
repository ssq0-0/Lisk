package process

import (
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/logger"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

var actionGenerators = map[string]func(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error){
	"Oku":                generateSwap,
	"IonicWithdrawAll":   generateIonicWithdraw,
	"IonicRepayAll":      generateIonicRepay,
	"Ionic15Borrow":      generate15Borrow(globals.Borrow, globals.LISK, globals.IonicBorrow),
	"Ionic71Supply":      generateIonic71Supply,
	"Relay":              generateBridgeToLisk,
	"Checker":            generateChecker,
	"Portal_daily_check": generateDailyCheck,
	"Portal_main_tasks":  generateMainTasks,
}

func generateTimeWindow(totalTime, actionCount int) []time.Duration {
	if actionCount <= 0 {
		return nil
	}

	baseInterval := time.Duration(totalTime) * time.Minute / time.Duration(actionCount)
	const variationFactor = 0.2
	const minVariation = time.Second
	const maxVariation = 30 * time.Second

	intervals := make([]time.Duration, 0, actionCount)
	for i := 0; i < actionCount; i++ {
		variation := float64(baseInterval) * variationFactor

		if variation < float64(minVariation) {
			variation = float64(minVariation)
		} else if variation > float64(maxVariation) {
			variation = float64(maxVariation)
		}

		randomVariation := time.Duration(rand.Float64()*2*variation - variation)
		interval := baseInterval + randomVariation

		if interval < 0 {
			interval = baseInterval
		}

		intervals = append(intervals, interval)
	}

	return intervals
}

func generateNextAction(acc *account.Account, selectedModule string, clients map[string]*ethClient.Client) (ActionProcess, error) {
	generator, exists := actionGenerators[selectedModule]
	if !exists {
		return ActionProcess{TypeAction: globals.Unknown}, fmt.Errorf("no action generator for module '%s'", selectedModule)
	}

	return generator(acc, clients)
}

func generateChecker(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	return packActionProcessStruct(globals.Checker, "Portal", big.NewInt(0), globals.NULL, globals.NULL), nil
}

func generateDailyCheck(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	return packActionProcessStruct(globals.DailyCheck, "Portal", big.NewInt(0), globals.NULL, globals.NULL), nil
}

func generateMainTasks(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	return packActionProcessStruct(globals.MainTasks, "Portal", big.NewInt(0), globals.NULL, globals.NULL), nil
}

func generateSwap(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	for attemps := 0; attemps < 5; attemps++ {
		if err := validateSwapHistory(acc.LastSwaps); err != nil {
			return ActionProcess{TypeAction: globals.Unknown}, err
		}

		tokenFrom := selectTokenFrom(acc)
		tokenTo := selectDifferentToken(tokenFrom)

		ethBal, err := validateNativeBalance(acc.Address, clients["lisk"])
		if err != nil {
			return ActionProcess{TypeAction: globals.Unknown}, err
		}

		if ethBal.Cmp(globals.MinBalances[globals.WETH]) < 0 {
			tokenTo = globals.WETH
			for attemps := 0; attemps < 5; attemps++ {
				tokenFrom = selectDifferentToken(tokenTo)
			}
		}

		amount, err := canDoActionByBalance(tokenFrom, acc, clients["lisk"])
		if err != nil {
			return ActionProcess{TypeAction: globals.Unknown}, err
		}

		updateSwapHistory(acc, tokenFrom, tokenTo)
		return packActionProcessStruct(globals.Swap, "Oku", amount, tokenFrom, tokenTo), nil
	}

	return ActionProcess{TypeAction: globals.Unknown}, fmt.Errorf("failed to generate swap after 5 attempts")
}

func generate15Borrow(actionType globals.ActionType, token common.Address, amount *big.Int) func(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	return func(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
		if actionType == globals.Borrow && acc.LiquidityState.ActionCount == 0 {
			acc.LiquidityState.ActionCount++
			return packActionProcessStruct(globals.EnterMarket, "Ionic", big.NewInt(0), globals.USDT, globals.NULL), nil
		}

		acc.LiquidityState.LastAction = actionType
		return packActionProcessStruct(actionType, "Ionic", amount, token, globals.NULL), nil
	}
}

func generateIonic71Supply(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	return packActionProcessStruct(globals.Supply, "Ionic", globals.IonicSupply, globals.USDT, globals.NULL), nil
}

func generateIonicRepay(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	return packActionProcessStruct(globals.Repay, "Ionic", globals.MaxRepayBigInt, globals.LISK, globals.NULL), nil
}

func generateIonicWithdraw(acc *account.Account, clients map[string]*ethClient.Client) (ActionProcess, error) {
	switch acc.LiquidityState.LastAction {
	case globals.ExitMarket:
		updateLiquidityState(acc, globals.Redeem)
		return packActionProcessStruct(globals.Redeem, "Ionic", globals.MaxRepayBigInt, globals.USDT, globals.NULL), nil
	default:
		updateLiquidityState(acc, globals.Repay)
		return packActionProcessStruct(globals.ExitMarket, "Ionic", globals.MaxRepayBigInt, globals.USDT, globals.NULL), nil
	}
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
