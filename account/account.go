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

	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"
)

type Account struct {
	Address             common.Address
	PrivateKey          *ecdsa.PrivateKey
	LastSwaps           []models.SwapPair
	LiquidityState      *models.LiquidityState
	Stats               int
	ActionsCount        int
	ActionsTime         int
	BalancePercentUsage int
	Proxy               string
}

func AccsFactory(privateKeys []string, cfg *config.Config) ([]*Account, error) {
	if len(privateKeys) == 0 {
		return nil, errors.New("privateKeys list is empty")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	var (
		accs []*Account
		mu   sync.Mutex
	)

	accs = make([]*Account, 0, len(privateKeys))

	for _, pk := range privateKeys {
		pk := pk

		g.Go(func() error {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			randSource := rand.New(rand.NewSource(time.Now().UnixNano()))

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

			account := &Account{
				Address:             publicAddr,
				PrivateKey:          privateKey,
				LastSwaps:           []models.SwapPair{},
				LiquidityState:      &models.LiquidityState{},
				ActionsCount:        cfg.ActionCounts,
				ActionsTime:         10 + randSource.Intn(25),
				BalancePercentUsage: 30 + randSource.Intn(47),
				Stats:               0,
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
