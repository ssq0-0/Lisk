package account

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"lisk/config"
	"lisk/globals"
	"lisk/models"
	"lisk/utils"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/sync/errgroup"
)

type Account struct {
	Address             common.Address
	PrivateKey          *ecdsa.PrivateKey
	RawPK               string
	LastSwaps           []models.SwapPair
	WrapHistory         models.WrapHistory
	WrapRange           models.WrapRange
	SwapRange           models.SwapRange
	LiquidityState      *models.LiquidityState
	Stats               map[string]int
	Mu                  sync.Mutex
	ActionsCount        int
	ActionsTime         int
	BalancePercentUsage int
	Proxy               string
}

func AccsFactory(privateKeys, proxys []string, cfg *config.Config, selectedModule string) ([]*Account, error) {
	if len(privateKeys) == 0 {
		return nil, errors.New("privateKeys list is empty")
	}

	if selectedModule == "AirdropStatus" {
		var accs []*Account
		for _, keyOrAddress := range privateKeys {
			var address common.Address

			// Проверяем, является ли переданное значение приватным ключом
			privKey, err := crypto.HexToECDSA(keyOrAddress)
			if err == nil {
				// Если это приватный ключ, извлекаем адрес из публичного ключа
				publicKey := privKey.Public().(*ecdsa.PublicKey)
				address = crypto.PubkeyToAddress(*publicKey)
			} else {
				// Если это не приватный ключ, трактуем его как обычный адрес
				address = common.HexToAddress(keyOrAddress)
			}

			accs = append(accs, &Account{
				Address:      address,
				ActionsCount: 1,
				Stats:        make(map[string]int),
			})
		}
		return accs, nil
	}

	wrapRange, swapRange, err := prepareRanges(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare ranges: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	var (
		accs []*Account
		mu   sync.Mutex
	)

	accs = make([]*Account, 0, len(privateKeys))

	for i, pk := range privateKeys {
		pk := pk

		g.Go(func() error {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			privateKey, err := utils.ParsePrivateKey(pk)
			if err != nil {
				return err
			}

			publicAddr, err := utils.DeriveAddress(privateKey)
			if err != nil {
				return err
			}

			if cfg.ActionCounts == 0 {
				return fmt.Errorf("0 actions count, check config")
			}

			var proxy string
			if len(proxys) > i {
				proxy = proxys[i]
			} else {
				proxy = ""
			}

			account := &Account{
				Address:             publicAddr,
				PrivateKey:          privateKey,
				RawPK:               pk,
				LastSwaps:           []models.SwapPair{},
				WrapHistory:         models.WrapHistory{},
				WrapRange:           *wrapRange,
				SwapRange:           *swapRange,
				LiquidityState:      &models.LiquidityState{},
				ActionsCount:        cfg.ActionCounts,
				ActionsTime:         cfg.MaxActionsTime,
				BalancePercentUsage: cfg.OkuPercentUsage,
				Stats:               make(map[string]int),
				Proxy:               proxy,
			}

			mu.Lock()
			accs = append(accs, account)
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return accs, nil
}

func prepareRanges(cfg *config.Config) (*models.WrapRange, *models.SwapRange, error) {
	var wrapRange models.WrapRange
	if err := utils.ConvertRangeAmount(cfg.WrapMinAmount, cfg.WrapMaxAmount, 18, globals.NULL, &wrapRange); err != nil {
		return nil, nil, err
	}

	swapRange := models.SwapRange{
		MinSwapAmount: make(map[common.Address]*big.Int),
		MaxSwapAmount: make(map[common.Address]*big.Int),
	}

	addresses := []common.Address{globals.USDT, globals.USDC, globals.WETH}
	for _, addr := range addresses {
		var err error
		switch addr {
		case globals.USDT, globals.USDC:
			err = utils.ConvertRangeAmount(cfg.SwapUSDTMinAmount, cfg.SwapUSDTMaxAmount, 6, addr, &swapRange)
		case globals.WETH:
			err = utils.ConvertRangeAmount(cfg.SwapEthMinAmount, cfg.SwapEthMaxAmount, 18, addr, &swapRange)
		}
		if err != nil {
			return &models.WrapRange{}, &models.SwapRange{}, err
		}
	}

	return &wrapRange, &swapRange, nil
}
