package common

import (
	"encoding/hex"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"

	hdwallet "github.com/dan13ram/go-ethereum-hdwallet"
)

const (
	AddressLength          = 20
	CosmosPublicKeyLength  = 33
	DefaultBIP39Passphrase = ""
	DefaultCosmosHDPath    = "m/44'/118'/0'/0/0"
	DefaultETHHDPath       = "m/44'/60'/0'/0/0"
)

var (
	defaultAlgo = hd.Secp256k1
)

func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

func isHexAddress(s string) bool {
	if has0xPrefix(s) {
		s = s[2:]
	}
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
	if has0xPrefix(s) {
		s = s[2:]
	}
	for _, c := range []byte(s) {
		if c != '0' {
			return false
		}
	}
	return true
}

func IsValidEthereumAddress(address string) bool {
	return has0xPrefix(address) && isHexAddress(address) && !isZeroHex(address)
}

func IsValidBech32Address(bech32Prefix string, address string) bool {
	addrCdc := addresscodec.NewBech32Codec(bech32Prefix)
	_, err := addrCdc.StringToBytes(address)
	return err == nil
}

func IsValidCosmosPublicKey(s string) bool {
	return !has0xPrefix(s) && isHex(s) && len(s) == 2*CosmosPublicKeyLength && !isZeroHex(s)
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

func EthereumAddressFromMnemonic(mnemonic string) (string, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)

	path := hdwallet.MustParseDerivationPath(DefaultETHHDPath)
	account, err := wallet.Derive(path, false)
	if err != nil {
		return "", err
	}

	return wallet.AddressHex(account)
}
