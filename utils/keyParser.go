package utils

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func ParsePrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	privateKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}

	return crypto.ToECDSA(privateKeyBytes)
}

func DeriveAddress(privateKey *ecdsa.PrivateKey) (common.Address, error) {
	if privateKey == nil {
		return (common.Address{}), fmt.Errorf("nil value in derive address function")
	}

	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	return crypto.PubkeyToAddress(*publicKey), nil
}
