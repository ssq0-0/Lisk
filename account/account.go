package account

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"lisk/config"
	"lisk/models"
	"lisk/utils"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"
)

type Account struct {
	Address             common.Address
	PrivateKey          *ecdsa.PrivateKey
	RawPK               string
	LastSwaps           []models.SwapPair
	WrapHistory         models.WrapHistory
	WrapRange           models.WrapRange
	LiquidityState      *models.LiquidityState
	Stats               map[string]int
	Mu                  sync.Mutex
	ActionsCount        int
	ActionsTime         int
	BalancePercentUsage int
	Proxy               string
}

func AccsFactory(privateKeys, proxys []string, cfg *config.Config) ([]*Account, error) {
	if len(privateKeys) == 0 {
		return nil, errors.New("privateKeys list is empty")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	wrapRange, err := utils.ConvertWrapAmount(cfg.WrapMinAmount, cfg.WrapMaxAmount)
	if err != nil {
		return nil, err
	}

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
				return fmt.Errorf("0 actions count. Check config.")
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
				WrapRange:           wrapRange,
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
