package util

import (
	"encoding/hex"
	"errors"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

const (
	DefaultEntropySize     = 256
	DefaultBIP39Passphrase = ""
	DefaultCosmosHDPath    = "m/44'/118'/0'/0/0"
)

var (
	DefaultAlgo = hd.Secp256k1
)

func AddressBytesFromBech32(bech32Prefix string, address string) (addr []byte, err error) {
	addrCdc := addresscodec.NewBech32Codec(bech32Prefix)
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
func Bech32FromAddressBytes(bech32Prefix string, bs []byte) (string, error) {
	if len(bs) == 0 {
		return "", nil
	}
	return bech32.ConvertAndEncode(bech32Prefix, bs)
}

func PrivKeyFromMnemonic(mnemonic string) (crypto.PrivKey, error) {

	// create master key and derive first key for keyring
	derivedPriv, err := DefaultAlgo.Derive()(mnemonic, DefaultBIP39Passphrase, DefaultCosmosHDPath)
	if err != nil {
		return nil, err
	}

	privKey := DefaultAlgo.Generate()(derivedPriv)

	return privKey, nil
}

func PubKeyFromHex(pubKeyHex string) (crypto.PubKey, error) {
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return nil, err
	}

	pubKey := &secp256k1.PubKey{}
	err = pubKey.UnmarshalAmino(pubKeyBytes)

	return pubKey, err
}
