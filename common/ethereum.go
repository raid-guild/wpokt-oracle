package common

import (
	"crypto/ecdsa"
	"strings"

	hdwallet "github.com/dan13ram/go-ethereum-hdwallet"
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
