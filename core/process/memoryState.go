package process

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type AccountState struct {
	AccountAddress  string `json:"address"`
	LastActionIndex int    `json:"last_action_index"`
	Module          string `json:"module"`
}

type Memory struct {
	StateFilePath string
	mu            sync.RWMutex
	states        map[string]*AccountState
}

func NewMemory(stateFilePath string) (*Memory, error) {
	m := &Memory{
		StateFilePath: stateFilePath,
		states:        make(map[string]*AccountState),
	}

	if err := m.loadFromFile(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Memory) loadFromFile() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	file, err := os.Open(m.StateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("не удалось открыть файл состояния: %w", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл состояния: %w", err)
	}

	if len(bytes) == 0 {
		return nil
	}

	var states []AccountState
	if err := json.Unmarshal(bytes, &states); err != nil {
		return fmt.Errorf("не удалось распарсить файл состояния: %w", err)
	}

	for _, state := range states {
		s := state
		m.states[state.AccountAddress] = &s
	}

	return nil
}

func (m *Memory) saveToFile() error {

	var states []AccountState
	for _, state := range m.states {
		states = append(states, *state)
	}

	bytes, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return fmt.Errorf("не удалось сериализовать состояния: %w", err)
	}

	if err := ioutil.WriteFile(m.StateFilePath, bytes, 0644); err != nil {
		return fmt.Errorf("не удалось записать файл состояния: %w", err)
	}

	return nil
}

func (m *Memory) LoadState(accountAddress string) (*AccountState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[accountAddress]
	if !exists {
		return nil, nil
	}
	return state, nil
}

func (m *Memory) IsStateFileNotEmpty() (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.states) > 0, nil
}

func (m *Memory) UpdateState(accountAddress string, module string, actionIndex int) error {
	m.mu.Lock()

	defer m.mu.Unlock()

	state, exists := m.states[accountAddress]
	if !exists {
		m.states[accountAddress] = &AccountState{
			AccountAddress:  accountAddress,
			LastActionIndex: actionIndex,
			Module:          module,
		}
	} else {
		state.LastActionIndex = actionIndex
		state.Module = module
	}

	return m.saveToFile()
}

func (m *Memory) ClearState(accountAddress string) error {
	m.mu.Lock()

	defer m.mu.Unlock()

	delete(m.states, accountAddress)
	return m.saveToFile()
}

func (m *Memory) ClearAllStates() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.states = make(map[string]*AccountState)
	return os.Truncate(m.StateFilePath, 0)
}
