package common

import (
	"crypto/ecdsa"
	"fmt"
	"strings"

	hdwallet "github.com/dan13ram/go-ethereum-hdwallet"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/crypto"
)

func IsValidEthereumAddress(s string) bool {
	s = strings.ToLower(s)
	_, err := BytesFromAddressHex(s)
	return err == nil && strings.HasPrefix(s, "0x")
}

func EthereumPrivateKeyFromMnemonic(mnemonic string) (*ecdsa.PrivateKey, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	path := hdwallet.MustParseDerivationPath(DefaultETHHDPath)
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}

	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func EthereumAddressFromMnemonic(mnemonic string) (string, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return "", err
	}

	path := hdwallet.MustParseDerivationPath(DefaultETHHDPath)
	account, err := wallet.Derive(path, false)
	if err != nil {
		return "", err
	}

	return wallet.AddressHex(account)
}

func HexToAddress(hex string) ethcommon.Address {
	return ethcommon.HexToAddress(hex)
}

func EthereumPrivateKeyToAddressHex(privateKey *ecdsa.PrivateKey) (string, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address.Hex(), nil
}
