package common

import (
	"crypto/ecdsa"
	"encoding/hex"
	"strings"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/types"

	hdwallet "github.com/dan13ram/go-ethereum-hdwallet"
)

const (
	AddressLength          = 20
	CosmosPublicKeyLength  = 33
	DefaultBIP39Passphrase = ""
	DefaultCosmosHDPath    = "m/44'/118'/0'/0/0"
	DefaultETHHDPath       = "m/44'/60'/0'/0/0"
	ZeroAddress            = "0x0000000000000000000000000000000000000000"
)

var (
	defaultAlgo = hd.Secp256k1
)

func isHexAddress(s string) bool {
	s = strings.ToLower(s)
	s = strings.TrimPrefix(s, "0x")
	return len(s) == 2*AddressLength && isHex(s)
}

func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

// isHex validates whether each byte is valid hexadecimal string.
func isHex(str string) bool {
	if len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}

func isZeroHex(s string) bool {
	s = strings.ToLower(s)
	s = strings.TrimPrefix(s, "0x")
	for _, c := range []byte(s) {
		if c != '0' {
			return false
		}
	}
	return true
}

func IsValidEthereumAddress(s string) bool {
	s = strings.ToLower(s)
	return strings.HasPrefix(s, "0x") && isHexAddress(s) && !isZeroHex(s)
}

func IsValidBech32Address(bech32Prefix string, address string) bool {
	addrCdc := addresscodec.NewBech32Codec(bech32Prefix)
	_, err := addrCdc.StringToBytes(address)
	return err == nil
}

func IsValidCosmosPublicKey(s string) bool {
	s = strings.ToLower(s)
	return !strings.HasPrefix(s, "0x") && isHex(s) && len(s) == 2*CosmosPublicKeyLength && !isZeroHex(s)
}

func CosmosPrivateKeyFromMnemonic(mnemonic string) (types.PrivKey, error) {
	derivedPriv, err := defaultAlgo.Derive()(mnemonic, DefaultBIP39Passphrase, DefaultCosmosHDPath)
	if err != nil {
		return nil, err
	}

	privKey := defaultAlgo.Generate()(derivedPriv)

	return privKey, nil
}

func CosmosPublicKeyFromMnemonic(mnemonic string) (string, error) {
	derivedPriv, err := defaultAlgo.Derive()(mnemonic, DefaultBIP39Passphrase, DefaultCosmosHDPath)
	if err != nil {
		return "", err
	}

	privKey := defaultAlgo.Generate()(derivedPriv)

	pubKey := privKey.PubKey()

	return hex.EncodeToString(pubKey.Bytes()), nil
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
