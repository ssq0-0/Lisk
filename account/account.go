package account

import (
	"crypto/ecdsa"
	"lisk/models"
	"lisk/utils"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Account struct {
	Address             common.Address
	PrivateKey          *ecdsa.PrivateKey
	LastSwaps           []models.SwapPair
	LiquidityState      *models.LiquidityState
	ActionsCount        int
	ActionsTime         int
	BalancePercentUsage int
	Proxy               string
}

func AccsFactory(privateKeys []string) ([]*Account, error) {
	var (
		accs  []*Account
		errCh = make(chan error)
		mu    sync.Mutex
		wg    sync.WaitGroup
	)

	rand.NewSource(time.Now().UnixNano())

	for _, pk := range privateKeys {
		wg.Add(1)

		go func(pk string) {
			defer wg.Done()
			privateKey, err := utils.ParsePrivateKey(pk)
			if err != nil {
				errCh <- err
				return
			}

			publicAddr, err := utils.DeriveAddress(privateKey)
			if err != nil {
				errCh <- err
				return
			}

			account := Account{
				Address:             publicAddr,
				PrivateKey:          privateKey,
				LastSwaps:           []models.SwapPair{},
				LiquidityState:      &models.LiquidityState{},
				ActionsCount:        3 + rand.Intn(25),
				ActionsTime:         10 + rand.Intn(25),
				BalancePercentUsage: 30 + rand.Intn(40),
			}

			mu.Lock()
			accs = append(accs, &account)
			mu.Unlock()
		}(pk)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return nil, <-errCh
	}

	return accs, nil
}
