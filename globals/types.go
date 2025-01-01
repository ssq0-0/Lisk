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
	DailyCheck     ActionType = "dailyCheck"
	MainTasks      ActionType = "mainTasks"
	HoldETH        ActionType = "holdETH"
	HoldLISK       ActionType = "holdLISK"
	HoldUSDT       ActionType = "holdUSDT"
	HoldUSDC       ActionType = "holdUSDC"
	HoldNFT        ActionType = "holdNFT"
	TwitterDiscord ActionType = "twitterDiscord"
	FonbnkVerif    ActionType = "fonbunkVerif"
	XelarkVerif    ActionType = "xelarVerif"
	Gitcoin        ActionType = "gitcoin"
)

var (
	Actions = []ActionType{Swap, Redeem, Supply, Borrow, Repay, EnterMarket, ExitMarket}
	Bridge  = map[string]ActionType{
		"base":     BaseBridge,
		"arbitrum": ArbitrumBridge,
		"optimism": OptimismBridge,
		"linea":    LineaBridge,
	}
	LiskPortalIDs = map[ActionType]map[ActionType]int{
		MainTasks: map[ActionType]int{
			HoldETH:        6,
			HoldLISK:       7,
			HoldUSDC:       8,
			HoldUSDT:       9,
			HoldNFT:        15,
			TwitterDiscord: 11,
			FonbnkVerif:    13,
			XelarkVerif:    14,
			Gitcoin:        12,
		},
		DailyCheck: map[ActionType]int{
			DailyCheck: 1,
		},
		Checker: map[ActionType]int{
			Checker: 100,
		},
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
