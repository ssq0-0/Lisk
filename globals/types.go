package globals

type ActionType string

const (
	Unknown        ActionType = "unknown"
	Swap           ActionType = "swap"
	Redeem         ActionType = "redeemUnderlying"
	Supply         ActionType = "supply"
	Borrow         ActionType = "borrow"
	Repay          ActionType = "repay"
	EnterMarket    ActionType = "enterMarkets"
	ExitMarket     ActionType = "exitMarket"
	ArbitrumBridge ActionType = "arbitrum"
	OptimismBridge ActionType = "optimism"
	LineaBridge    ActionType = "linea"
	BaseBridge     ActionType = "base"
	Checker        ActionType = "checker"
)

var (
	Actions = []ActionType{Swap, Redeem, Supply, Borrow, Repay, EnterMarket, ExitMarket}
	Bridge  = map[string]ActionType{
		"base":     BaseBridge,
		"arbitrum": ArbitrumBridge,
		"optimism": OptimismBridge,
		"linea":    LineaBridge,
	}
)

var (
	SwapIn    = []byte{0x00}
	WrapETH   = []byte{0x0b}
	UnwrapETH = []byte{0x0c}
)

type PoolInfo struct {
	Token0 int
	Token1 int
}
