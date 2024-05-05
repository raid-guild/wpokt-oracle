package utils

import (
	"encoding/hex"
	"errors"
	"fmt"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/go-bip39"
)

const (
	DefaultEntropySize     = 256
	DefaultBIP39Passphrase = ""
	DefaultCosmosHDPath    = "m/44'/118'/0'/0/0"
	DefaultBech32Prefix    = "pokt"

	DefaultETHHDPath = "m/44'/60'/0'/0/0"
)

var (
	DefaultAlgo = hd.Secp256k1
)

func AddressBytesFromBech32(address string) (addr []byte, err error) {

	addrCdc := addresscodec.NewBech32Codec(DefaultBech32Prefix)
	return addrCdc.StringToBytes(address)
}

func AddressBytesFromHexString(address string) ([]byte, error) {
	if len(address) == 0 {
		return nil, errors.New("decoding address from hex string failed: empty address")
	}

	if address[0:2] == "0x" || address[0:2] == "0X" {
		address = address[2:]
	}

	return hex.DecodeString(address)
}

// Bech32FromAddressBytes returns a bech32 representation of address bytes.
// Returns an empty string if the byte slice is 0-length. Returns an error if the bech32 conversion
// fails or the prefix is empty.
func Bech32FromAddressBytes(bs []byte) (string, error) {
	if len(bs) == 0 {
		return "", nil
	}
	return bech32.ConvertAndEncode(DefaultBech32Prefix, bs)
}

func NewMnemonic() (string, error) {
	// Default number of words (24): This generates a mnemonic directly from the
	// number of words by reading system entropy.
	entropy, err := bip39.NewEntropy(DefaultEntropySize)
	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
}

func NewAccount(mnemonic string) (crypto.PrivKey, error) {

	// create master key and derive first key for keyring
	derivedPriv, err := DefaultAlgo.Derive()(mnemonic, DefaultBIP39Passphrase, DefaultCosmosHDPath)
	if err != nil {
		return nil, err
	}

	privKey := DefaultAlgo.Generate()(derivedPriv)

	return privKey, nil
}

func TestAddress(bech32Addr string) {
	if bech32Addr == "" {
		bech32Addr = "pokt1mrqt5f7qh8uxs27cjm9t7v9e74a9vvdnq5jva4"
	}

	fmt.Printf("Bech32 Address: %s\n", bech32Addr)

	accAddr, _ := AddressBytesFromBech32(bech32Addr)

	fmt.Println("Bytes: ", accAddr)
	fmt.Printf("Length: %d\n", len(accAddr))

	fmt.Printf("Hex: 0x%s\n", hex.EncodeToString(accAddr))

	bytes, _ := AddressBytesFromHexString(hex.EncodeToString(accAddr))

	fmt.Println("Bytes: ", bytes)
	fmt.Printf("Length: %d\n", len(bytes))

	output, _ := Bech32FromAddressBytes(bytes)

	fmt.Println("Bech32: ", output)

	isEqual := false
	for i := range accAddr {
		if accAddr[i] != bytes[i] {
			isEqual = false
			break
		}
		isEqual = true
	}

	fmt.Println("Are bytes equal: ", isEqual)
	fmt.Println("Are bech32 equal: ", bech32Addr == output)
}
