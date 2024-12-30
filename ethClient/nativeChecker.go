package ethClient

import (
	"lisk/globals"

	"github.com/ethereum/go-ethereum/common"
)

func IsNativeToken(tokenAddr common.Address) bool {
	nativeTokens := map[common.Address]bool{
		globals.WETH: true,
	}

	return nativeTokens[tokenAddr]
}
