package util

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultEntropySize     = 256
	DefaultBIP39Passphrase = ""
	DefaultCosmosHDPath    = "m/44'/118'/0'/0/0"
)

var (
	DefaultAlgo = hd.Secp256k1
)

func BytesFromBech32(bech32Prefix string, address string) (addr []byte, err error) {
	addrCdc := addresscodec.NewBech32Codec(bech32Prefix)
	return addrCdc.StringToBytes(address)
}

func BytesFromHex(hexStr string) ([]byte, error) {
	if len(hexStr) == 0 {
		return nil, errors.New("decoding hex string failed: empty hex string")
	}
	hexStr = strings.ToLower(hexStr)
	hexStr = strings.TrimPrefix(hexStr, "0x")
	return hex.DecodeString(hexStr)
}

func Ensure0xPrefix(str string) string {
	str = strings.ToLower(str)
	if !strings.HasPrefix(str, "0x") {
		return "0x" + str
	}
	return str
}

func HexFromBytes(address []byte) string {
	return Ensure0xPrefix(hex.EncodeToString(address))
}

func Bech32FromBytes(bech32Prefix string, bs []byte) (string, error) {
	if len(bs) == 0 {
		return "", nil
	}
	return bech32.ConvertAndEncode(bech32Prefix, bs)
}

func PrivKeyFromMnemonic(mnemonic string) (crypto.PrivKey, error) {
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

func CoinsToBigInt(coins sdk.Coins) (*big.Int, error) {
	if coins.Len() == 0 {
		return big.NewInt(0), nil
	}

	if coins.Len() > 1 {
		return big.NewInt(0), fmt.Errorf("coins contain more than one type of coin")
	}

	coin := coins[0] // Get the first coin in the Coins array
	amountStr := coin.Amount.String()
	bigIntAmount := new(big.Int)
	bigIntAmount, ok := bigIntAmount.SetString(amountStr, 10)
	if !ok {
		return big.NewInt(0), fmt.Errorf("unable to convert coin amount to big.Int")
	}

	return bigIntAmount, nil
}
