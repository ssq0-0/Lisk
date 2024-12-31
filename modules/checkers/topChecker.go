package checkers

import (
	"fmt"
	"lisk/account"
	"lisk/globals"
	"lisk/httpClient"
	"lisk/logger"
	"lisk/models"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Checker struct {
	Endpoint string
}

func NewChecker(endpoint string) (*Checker, error) {
	return &Checker{
		Endpoint: endpoint,
	}, nil
}

func (c *Checker) Action(tokenIn, tokenOut common.Address, amountIn *big.Int, acc *account.Account, ta globals.ActionType) error {
	client, err := httpClient.NewHttpClient(acc.Proxy)
	if err != nil {
		return nil
	}

	request, err := c.createRequest(acc.Address)
	if err != nil {
		return nil
	}

	var result models.GraphQLResponse
	if err := client.SendJSONRequest(c.Endpoint, "POST", request, &result); err != nil {
		return err
	}

	if len(result.Errors) > 0 {
		fmt.Printf("GraphQL Errors: %+v\n", result.Errors)
		return fmt.Errorf("GraphQL errors: %v", result.Errors)
	}

	rank := result.Data.Userdrop.User.Rank
	points := result.Data.Userdrop.User.Points
	updateAt := result.Data.Userdrop.User.UpdateAt

	logger.GlobalLogger.Infof("Statics for %s: Rank: %d | Points: %d | Last Top Update: %s", acc.Address.Hex(), rank, points, updateAt)
	return nil
}

func (c *Checker) createRequest(address common.Address) (models.GraphQLRequest, error) {
	query := `
	query AirdropUser($filter: UserFilter!, $pointsHistoryFilter: QueryFilter, $tasksFilter: QueryFilter) {
		userdrop {
			user(filter: $filter) {
				rank
				points
				updatedAt
				pointsHistories(filter: $pointsHistoryFilter) {
					totalCount
				}
				tasks(filter: $tasksFilter) {
					id
				}
			}
		}
	}
`

	variables := map[string]interface{}{
		"filter": map[string]interface{}{
			"address": address.Hex(),
		},
		"pointsHistoryFilter": map[string]interface{}{},
		"tasksFilter":         map[string]interface{}{},
	}

	gqlReq := models.GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	return gqlReq, nil
}
