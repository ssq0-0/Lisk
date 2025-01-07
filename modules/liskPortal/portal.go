package liskPortal

import (
	"fmt"
	"lisk/account"
	"lisk/globals"
	"lisk/httpClient"
	"lisk/logger"
	"lisk/models"
	"lisk/utils"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Portal struct {
	TaskEndpoint    string
	CheckerEndpoint string
}

func NewPortal(taskEndpoint, checkerEndpoint string) (*Portal, error) {
	if taskEndpoint == "" || checkerEndpoint == "" {
		return nil, fmt.Errorf("endpoint lisk portal is empty, check config.")
	}

	return &Portal{
		TaskEndpoint:    taskEndpoint,
		CheckerEndpoint: checkerEndpoint,
	}, nil
}

func (p *Portal) Action(tokenIn, tokenOut common.Address, amountIn *big.Int, acc *account.Account, ta globals.ActionType) error {
	client, err := httpClient.NewHttpClient(acc.Proxy)
	if err != nil {
		return err
	}

	switch ta {
	case globals.DailyCheck:
		return p.handleSingleTask(client, acc, ta, p.TaskEndpoint)

	case globals.MainTasks:
		return p.handleMultipleTasks(client, acc)

	default:
		return p.handleSingleTask(client, acc, ta, p.CheckerEndpoint)
	}
}

func (p *Portal) handleSingleTask(client *httpClient.HttpClient, acc *account.Account, ta globals.ActionType, endpoint string) error {
	taskID, exists := globals.LiskPortalIDs[ta][ta]
	if !exists {
		return fmt.Errorf("task ID not found for action type: %s", ta)
	}

	request, err := p.createRequest(acc.Address, ta, taskID)
	if err != nil {
		return err
	}

	var result models.TaskResponse
	if err := client.SendJSONRequest(endpoint, "POST", request, &result); err != nil {
		return fmt.Errorf("failed to execute task %d: %w", taskID, err)
	}

	return p.processTaskResult(taskID, result, ta, acc.Address)
}

func (p *Portal) handleMultipleTasks(client *httpClient.HttpClient, acc *account.Account) error {
	taskMap, exists := globals.LiskPortalIDs[globals.MainTasks]
	if !exists {
		return fmt.Errorf("no tasks found for MainTasks")
	}

	for subTaskType, taskID := range taskMap {
		request, err := p.createRequest(acc.Address, subTaskType, taskID)
		if err != nil {
			logger.GlobalLogger.Errorf("Failed to create request for task %d: %v", taskID, err)
			continue
		}

		var result models.TaskResponse
		if err := client.SendJSONRequest(p.TaskEndpoint, "POST", request, &result); err != nil {
			logger.GlobalLogger.Errorf("Failed to execute task %d: %v", taskID, err)
			continue
		}

		p.processTaskResult(taskID, result, subTaskType, acc.Address)
		time.Sleep(time.Second * 1)
	}

	return nil
}

func (p *Portal) createRequest(address common.Address, ta globals.ActionType, taskID int) (models.GraphQLRequest, error) {
	switch ta {
	case globals.Gitcoin, globals.HoldETH, globals.HoldLISK, globals.HoldUSDC, globals.HoldUSDT, globals.TwitterDiscord,
		globals.HoldNFT, globals.FonbnkVerif, globals.XelarkVerif, globals.MainTasks, globals.DailyCheck:
		return models.GraphQLRequest{
			Query: TaskQuery,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"address": address.Hex(),
					"taskID":  taskID,
				},
			},
		}, nil
	case globals.Checker:
		return models.GraphQLRequest{
			Query: TopQuery,
			Variables: map[string]interface{}{
				"filter": map[string]interface{}{
					"address": address.Hex(),
				},
			},
		}, nil

	default:
		return models.GraphQLRequest{}, fmt.Errorf("unsupported action type: %s", ta)
	}
}

func (p *Portal) processTaskResult(taskID int, result models.TaskResponse, ta globals.ActionType, address common.Address) error {
	switch ta {
	case globals.Checker:
		userData := result.Data.Userdrop.User
		line := fmt.Sprintf("%s,%d,%d,%s", address.Hex(), userData.Rank, userData.Points, userData.UpdateAt)
		if err := utils.AppendLinesToFile(utils.GetPath("task_results"), []string{line}); err != nil {
			return fmt.Errorf("failed to append task result to file: %w", err)
		}
	case globals.DailyCheck, globals.MainTasks, globals.FonbnkVerif, globals.XelarkVerif, globals.Gitcoin, globals.HoldETH,
		globals.HoldLISK, globals.HoldUSDC, globals.HoldUSDT, globals.HoldNFT, globals.TwitterDiscord:
		taskStatus := result.Data.Userdrop.UpdateTaskStatus
		if !taskStatus.Success {
			logger.GlobalLogger.Errorf("[Failed %s] to for task %d: message: %v", address.Hex(), taskID, result.Errors)
		} else {
			logger.GlobalLogger.Infof("[Task %s] %d completed successfully", address.Hex(), taskID)
		}
	default:
		return fmt.Errorf("unsupported action type: %s", ta)
	}
	return nil
}
