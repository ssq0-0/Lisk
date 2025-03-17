package globals

import (
	"bytes"
	"lisk/logger"
	"time"

	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func init() {
	parsedABI, err := abi.JSON(bytes.NewReader(Erc20JSON))
	if err != nil {
		logger.GlobalLogger.Fatalf("Failed parsing ABI: %v", err)
	}

	Erc20ABI = &parsedABI
	_, success := MaxUint256.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10)
	if !success {
		logger.GlobalLogger.Fatalf("Failed to set MaxRepayBigInt: invalid number")
	}
}

const (
	SoftVersion  = "v2.9.0"
	LinkRepo     = "https://api.github.com/repos/ssq0-0/Lisk/releases/latest"
	Format       = "02.01.2006"
	TotalSuccess = 0
	TodaySuccess = 1
	ConsoleTitle = "Lisk | cheif.ssq"
	Blockscout   = "https://blockscout.lisk.com/api/v2/stats"
)

var (
	// StartDate sets the start date for calculating the week number.
	// Date format: time.Time
	// Example: 1 January 2025, time 00:00:00 UTC
	// Otherwise: 2025.01.01 00:00:00
	StartDate time.Time

	// LISK have 18 decimals => 18 '0' after ' , '
	// You can write 1e18 (1 LISK) or 1e17(0.1 LISK). Example 0.15LISK = 15e16. (15 + 16 zero)
	// min amount for borrow +-0.13 LISK. You can change this, but keep in mind the minimum amount
	IonicBorrow *big.Int

	// USDT/USDC have 6 decimal => 6 '0' after ' , '
	// Example 1 USDC => 1_000_000.
	// Example 1.1 USDC => 11_000_00
	// For a successful supply/borrow/return cycle, you need to make supply at least 66% more than you want to reciprocate.
	IonicSupply *big.Int // 0,09 USDT

	// Need for gas in tx. If ETH < MinETHForTx - the execution of the count will end as a whole
	MinETHForTx = big.NewInt(1e13) // 0.00001.

	// Need for gas in tx. If ETH < MinETHForTx - the execution of the count will end as a whole
	MinUsdtForTx big.Int

	// need for oku swaps config percent use
	// OkuPercentUsage int // default 50%

	// need for wrap/unwrap
	// WrapAmount *big.Int // set in config

	// Number of simultaneously operating goroutines for user control over the semaphore
	GorutinesCount int

	AttentionGwei    *big.Int // GWEI have 9 decimals
	AttentionTime    int      // Time in seconds that indicates how often to check the throttle reduction
	MaxAttentionTime int      //Time in minutes how long the waiting cycle will last at most, after which the execution will continue

	Slippage              = big.NewFloat(0.01) // 1%
	DefaultDeadlineOffset = 120                // 2 minutes
	ApproveDeadlineOffset = 3600               // 1 hour
	MaxApprove            = big.NewInt(1e18)
	Erc20ABI              *abi.ABI
	MaxUint256            = new(big.Int)

	WETH   = common.HexToAddress("0x4200000000000000000000000000000000000006")
	LISK   = common.HexToAddress("0xac485391EB2d7D88253a7F1eF18C37f4242D1A24")
	USDC   = common.HexToAddress("0xF242275d3a6527d877f2c927a82D9b057609cc71")
	USDT   = common.HexToAddress("0x05D032ac25d322df992303dCa074EE7392C117b9")
	NATIVE = common.Address{}
	NULL   = common.Address{} // need for minor functions

	// tokens array. Need for create tokens graph in pool algoritm
	Tokens = []common.Address{WETH, USDC, USDT, LISK}

	TokensDecimals = map[common.Address]PoolInfo{
		common.HexToAddress("0x87D3d9CA455DCc9a3Ba5605D2829d994922DD04F"): PoolInfo{Token0: 6, Token1: 6},
		common.HexToAddress("0xA211813F9d68F9c33bc9C275e40DfC4027016232"): PoolInfo{Token0: 6, Token1: 6},
		common.HexToAddress("0xA211813F9d68F9c33bc9C275e40DfC4027016232"): PoolInfo{Token0: 6, Token1: 6},
		common.HexToAddress("0x49DA5d42091F38fc527b9F9B03C1005aBb6aD818"): PoolInfo{Token0: 18, Token1: 6},
		common.HexToAddress("0x3A670179BdecE7eB4f570e30Ee9D560f7ff4Fac3"): PoolInfo{Token0: 18, Token1: 6},
		common.HexToAddress("0x0501f71EED6c1F9c1337823C1a48a1390D16235a"): PoolInfo{Token0: 18, Token1: 18},
		common.HexToAddress("0xD501d4E381491F64274Cc65fdec32b47264a2422"): PoolInfo{Token0: 18, Token1: 18},
	}

	DecimalsMap = map[common.Address]int{
		WETH: 18,
		LISK: 18,
		USDC: 6,
		USDT: 6,
	}

	MinBalances = map[common.Address]*big.Int{
		WETH: big.NewInt(1e14), // 0.00001
		USDT: &MinUsdtForTx,    // 0.1
		USDC: &MinUsdtForTx,    // 0.1
		LISK: big.NewInt(1e17), // 0.1
	}

	LimitedModules = map[string]int{
		"Portal_daily_check": 1,
		"Portal_main_tasks":  1,
		"Checker":            1,
		"BalanceCheck":       1,
		"Relay":              1,
		"IonicRepayAll":      1,
		"IonicWithdrawAll":   2,
		"Ionic71Supply":      72,
		"Ionic15Borrow":      15,
	}
)

var (
	Erc20JSON = []byte(`[
	{
		
		"constant":true,
		"inputs":[{"name":"account","type":"address"}],
		"name":"balanceOf",
		"outputs":[{"name":"","type":"uint256"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[{"name":"spender","type":"address"},{"name":"owner","type":"address"}],
		"name":"allowance",
		"outputs":[{"name":"","type":"uint256"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant": false,
		"inputs": [],
		"name": "deposit",
		"outputs": [],
		"payable": true,
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
			"internalType": "uint256",
			"name": "wad",
			"type": "uint256"
			}
		],
		"name": "withdraw",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant":false,
		"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],
		"name":"approve",
		"outputs":[{"name":"","type":"bool"}],
		"payable":false,
		"stateMutability":"nonpayable",
		"type":"function"
	},
	{
		"constant":false,
		"inputs":[{"name":"recipient","type":"address"},{"name":"amount","type":"uint256"}],
		"name":"transfer",
		"outputs":[{"name":"","type":"bool"}],
		"payable":false,
		"stateMutability":"nonpayable",
		"type":"function"
	},
	{
		"constant":false,
		"inputs":[{"name":"sender","type":"address"},{"name":"recipient","type":"address"},{"name":"amount","type":"uint256"}],
		"name":"transferFrom",
		"outputs":[{"name":"","type":"bool"}],
		"payable":false,
		"stateMutability":"nonpayable",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"decimals",
		"outputs":[{"name":"","type":"uint8"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"name",
		"outputs":[{"name":"","type":"string"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"symbol",
		"outputs":[{"name":"","type":"string"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"totalSupply",
		"outputs":[{"name":"","type":"uint256"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	}
]`)
)

// List of random headers and user agents to form a plausible http source request
var (
	SecChUa = map[string]string{
		"Macintosh": `Not)A;Brand";v="99", "Google Chrome";v="120", "Chromium";v="120"`,
		"Windows":   `"Not)A;Brand";v="99", "Google Chrome";v="120", "Chromium";v="120"`,
		"Linux":     `"Not)A;Brand";v="99", "Google Chrome";v="120", "Chromium";v="120"`,
		"Unknown":   `"Not)A;Brand";v="99", "Google Chrome";v="120", "Chromium";v="120"`,
	}

	Platforms = map[string]string{
		"Macintosh": `"macOS"`,
		"Windows":   `"Windows"`,
		"Linux":     `"Linux"`,
	}

	UserAgents = []string{
		// Google Chrome
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 12_6_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 12.6; rv:115.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:115.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.90 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.90 Safari/537.36",

		// Mozilla Firefox
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:115.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 12.6; rv:115.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:115.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:114.0) Gecko/20100101 Firefox/114.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 12.6; rv:114.0) Gecko/20100101 Firefox/114.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:114.0) Gecko/20100101 Firefox/114.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:113.0) Gecko/20100101 Firefox/113.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 12.6; rv:113.0) Gecko/20100101 Firefox/113.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:113.0) Gecko/20100101 Firefox/113.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:112.0) Gecko/20100101 Firefox/112.0",

		// Safari
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_1_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Safari/605.1.15",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 16_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPad; CPU OS 16_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 12.6; rv:115.0) Gecko/20100101 Firefox/115.0 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 12.6; rv:114.0) Gecko/20100101 Firefox/114.0 Safari/605.1.15",

		// Microsoft Edge
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 12.6; rv:115.0) Gecko/20100101 Firefox/115.0 Edg/120.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:115.0) Gecko/20100101 Firefox/115.0 Edg/120.0.0.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.90 Safari/537.36 Edg/118.0.5993.90",

		// Opera
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 12.6; rv:115.0) Gecko/20100101 Firefox/115.0 OPR/106.0.0.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:115.0) Gecko/20100101 Firefox/115.0 OPR/106.0.0.0",

		// Mobile Browsers
		"Mozilla/5.0 (iPhone; CPU iPhone OS 16_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPad; CPU OS 16_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 13; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; Pixel 7 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; SM-N975F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.90 Mobile Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 15_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPad; CPU OS 15_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 14; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 14; Pixel 8 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; SM-N975F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.90 Mobile Safari/537.36",

		// Additional Variations
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:115.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 11_6_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 11.6; rv:115.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:115.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64; rv:114.0) Gecko/20100101 Firefox/114.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Safari/605.1.15",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 16_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.4 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 14; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 14; Pixel 8 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:114.0) Gecko/20100101 Firefox/114.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:113.0) Gecko/20100101 Firefox/113.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.4 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7; rv:114.0) Gecko/20100101 Firefox/114.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:113.0) Gecko/20100101 Firefox/113.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.5993.90 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:112.0) Gecko/20100101 Firefox/112.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.0; rv:112.0) Gecko/20100101 Firefox/112.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:112.0) Gecko/20100101 Firefox/112.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:111.0) Gecko/20100101 Firefox/111.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.4 Safari/605.1.15",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 16_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.3 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPad; CPU OS 16_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.3 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 14; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 14; Pixel 8 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 12.6; rv:115.0) Gecko/20100101 Firefox/115.0 OPR/106.0.0.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:115.0) Gecko/20100101 Firefox/115.0 OPR/106.0.0.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 11.6.0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Safari/605.1.15",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 16_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.2 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPad; CPU OS 16_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.2 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 13; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; Pixel 7 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:114.0) Gecko/20100101 Firefox/114.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.0; rv:114.0) Gecko/20100101 Firefox/114.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:114.0) Gecko/20100101 Firefox/114.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:113.0) Gecko/20100101 Firefox/113.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.0; rv:113.0) Gecko/20100101 Firefox/113.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:113.0) Gecko/20100101 Firefox/113.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:112.0) Gecko/20100101 Firefox/112.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.0; rv:112.0) Gecko/20100101 Firefox/112.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:112.0) Gecko/20100101 Firefox/112.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:111.0) Gecko/20100101 Firefox/111.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.0; rv:111.0) Gecko/20100101 Firefox/111.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:111.0) Gecko/20100101 Firefox/111.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.5993.90 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.0; rv:115.0) Gecko/20100101 Firefox/115.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPad; CPU OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 12; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.90 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Pixel 6 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5993.90 Mobile Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 12.6; rv:115.0) Gecko/20100101 Firefox/115.0",
	}
)
