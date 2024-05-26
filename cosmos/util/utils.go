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
	"github.com/dan13ram/wpokt-oracle/models"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
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

// CoinsToBigInt converts sdk.Coins to big.Int
func CoinsToBigInt(coins sdk.Coins) (*big.Int, error) {
	// For simplicity, assume coins contain only one type of coin
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

func ParseMessageSenderEvent(
	events []abci.Event,
) (string, error) {
	for _, event := range events {
		if strings.EqualFold(event.Type, "message") {
			for _, attr := range event.Attributes {
				if strings.EqualFold(string(attr.Key), "sender") {
					sender := string(attr.Value)
					return sender, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no sender found in message events")
}

func ParseCoinsReceivedEvents(
	denom string,
	receiver string, events []abci.Event,
) (sdk.Coin, error) {
	total := sdk.NewCoin(denom, math.NewInt(0))
	for _, event := range events {
		if strings.EqualFold(event.Type, "coin_received") {
			for _, attr := range event.Attributes {
				if strings.EqualFold(string(attr.Key), "receiver") && strings.EqualFold(string(attr.Value), receiver) {
					for _, attr := range event.Attributes {
						if strings.EqualFold(string(attr.Key), "amount") {
							amountStr := string(attr.Value)
							amount, err := sdk.ParseCoinNormalized(amountStr)
							if err != nil {
								return total, fmt.Errorf("unable to parse coin amount: %v", err)
							}
							if amount.Denom != denom {
								return total, fmt.Errorf("invalid coin denom: %s", amount.Denom)
							}
							total = total.Add(amount)
						}
					}
				}
			}
		}
	}
	return total, nil
}

func ParseCoinsSpentEvents(
	denom string,
	events []abci.Event,
) (string, sdk.Coin, error) {
	total := sdk.NewCoin(denom, math.NewInt(0))
	spender := ""
	for _, event := range events {
		if strings.EqualFold(event.Type, "coin_spent") {
			for _, attr := range event.Attributes {
				if strings.EqualFold(string(attr.Key), "spender") {
					newSpender := string(attr.Value)
					if spender != "" && !strings.EqualFold(spender, newSpender) {
						return spender, total, fmt.Errorf("multiple spenders found in coin spent events")
					}
					spender = newSpender
				}
				if strings.EqualFold(string(attr.Key), "amount") {
					amountStr := string(attr.Value)
					amount, err := sdk.ParseCoinNormalized(amountStr)
					if err != nil {
						return spender, total, err
					}
					if amount.Denom != denom {
						return spender, total, fmt.Errorf("invalid coin denom: %s", amount.Denom)
					}
					total = total.Add(amount)
				}
			}
		}
	}
	return spender, total, nil
}

// hash the string chainID to get a uint64 chainDomain
func getChainDomain(chainID string) uint64 {
	chainHash := ethcrypto.Keccak256([]byte(chainID))
	chainDomain := new(big.Int).SetBytes(chainHash).Uint64()
	return chainDomain
}

func ParseChain(config models.CosmosNetworkConfig) models.Chain {
	return models.Chain{
		ChainID:     config.ChainID,
		ChainDomain: getChainDomain(config.ChainID),
		ChainName:   config.ChainName,
		ChainType:   models.ChainTypeCosmos,
	}
}
