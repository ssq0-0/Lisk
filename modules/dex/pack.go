package dex

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func (d *Dex) packSwapData(recipient common.Address, amountIn, amountOutMinimum *big.Int, path []byte, payFromMsgSender bool) ([]byte, error) {
	args := abi.Arguments{
		{Type: abi.Type{T: abi.AddressTy}},
		{Type: abi.Type{T: abi.UintTy, Size: 256}},
		{Type: abi.Type{T: abi.UintTy, Size: 256}},
		{Type: abi.Type{T: abi.BytesTy}},
		{Type: abi.Type{T: abi.BoolTy}},
	}
	return args.Pack(recipient, amountIn, amountOutMinimum, path, payFromMsgSender)
}

func (d *Dex) packWrapETHData(recipient common.Address, amountIn *big.Int) ([]byte, error) {
	args := abi.Arguments{
		{Type: abi.Type{T: abi.AddressTy}},         // recipient
		{Type: abi.Type{T: abi.UintTy, Size: 256}}, // amount
	}
	return args.Pack(recipient, amountIn)
}

func (d *Dex) encodeV3Path(tokenIn common.Address, fee *big.Int, tokenOut common.Address) ([]byte, error) {
	if fee.Cmp(big.NewInt(0xFFFFFF)) > 0 {
		return nil, fmt.Errorf("fee exceeds 24 bits")
	}
	feeUint32 := uint32(fee.Uint64())

	path := make([]byte, 0, 20+3+20)
	path = append(path, tokenIn.Bytes()...)

	feeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(feeBytes, feeUint32)

	path = append(path, feeBytes[1:4]...)
	path = append(path, tokenOut.Bytes()...)
	return path, nil
}
