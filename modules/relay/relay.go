package relay

import (
	"encoding/hex"
	"fmt"
	"lisk/account"
	"lisk/ethClient"
	"lisk/globals"
	"lisk/httpClient"
	"lisk/models"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type Relay struct {
	Client   map[globals.ActionType]*ethClient.Client
	Endpoint string
}

func NewRelay(clients map[globals.ActionType]*ethClient.Client, endpoint string) (*Relay, error) {
	return &Relay{Client: clients, Endpoint: endpoint}, nil
}

func (r *Relay) Action(tokenIn, tokenOut common.Address, amountIn *big.Int, acc *account.Account, ta globals.ActionType) error {
	client, err := httpClient.NewHttpClient(acc.Proxy)
	if err != nil {
		return err
	}

	chainID, err := r.Client[ta].GetChainID()
	if err != nil {
		return err
	}

	quoteData, err := r.getQuoteData(tokenIn, tokenOut, amountIn, int(chainID), acc, client)
	if err != nil {
		return err
	}

	to, value, data, err := r.prepareData(quoteData)
	if err != nil {
		return err
	}

	return r.Client[ta].SendTransaction(acc.PrivateKey, acc.Address, to, r.Client[ta].GetNonce(acc.Address), value, data)
}

func (r *Relay) getQuoteData(tokenIn, tokenOut common.Address, amountIn *big.Int, chainID int, acc *account.Account, client *httpClient.HttpClient) (*models.RelayResponse, error) {
	request := models.RelayRequest{
		User:                 acc.Address.Hex(),
		OriginChainId:        chainID,
		DestinationChainId:   1135,
		OriginCurrency:       tokenIn.Hex(),
		DestinationCurrency:  tokenOut.Hex(),
		Recipient:            acc.Address.Hex(),
		TradeType:            "EXACT_INPUT",
		Amount:               amountIn.String(),
		Referrer:             "relay.link/swap",
		UseExternalLiquidity: false,
		UseDepositAddress:    false,
	}

	var result models.RelayResponse
	if err := client.SendJSONRequest(r.Endpoint, "POST", request, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *Relay) prepareData(quoteData *models.RelayResponse) (common.Address, *big.Int, []byte, error) {
	to := quoteData.Steps[0].Items[0].Data.To
	data := quoteData.Steps[0].Items[0].Data.Data
	value := quoteData.Steps[0].Items[0].Data.Value

	valueBigInt, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return common.Address{}, nil, nil, fmt.Errorf("failed to convert value to big.Int: %s", value)
	}

	dataBytes, err := hex.DecodeString(strings.TrimPrefix(data, "0x"))
	if err != nil {
		return common.Address{}, nil, nil, fmt.Errorf("failed to decode data: %w", err)
	}

	toAddress := common.HexToAddress(to)

	return toAddress, valueBigInt, dataBytes, nil
}
