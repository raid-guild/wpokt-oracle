package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"strings"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/go-bip39"

	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

const (
	defaultEntropySize     = 256
	defaultBIP39Passphrase = ""
	defaultCosmosHDPath    = "m/44'/118'/0'/0/0"
	defaultBech32Prefix    = "pokt"

	defaultETHHDPath = "m/44'/60'/0'/0/0"
)

var (
	defaultAlgo = hd.Secp256k1
)

func AddressBytesFromBech32(address string) (addr []byte, err error) {

	addrCdc := addresscodec.NewBech32Codec(defaultBech32Prefix)
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
	return bech32.ConvertAndEncode(defaultBech32Prefix, bs)
}

func NewMnemonic() (string, error) {
	// Default number of words (24): This generates a mnemonic directly from the
	// number of words by reading system entropy.
	entropy, err := bip39.NewEntropy(defaultEntropySize)
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
	derivedPriv, err := defaultAlgo.Derive()(mnemonic, defaultBIP39Passphrase, defaultCosmosHDPath)
	if err != nil {
		return nil, err
	}

	privKey := defaultAlgo.Generate()(derivedPriv)

	return privKey, nil
}

func testAddress(bech32Addr string) {
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

func main() {

	var mnemonic string
	var err error
	flag.StringVar(&mnemonic, "mnemonic", "", "24 word mnemonic")
	flag.Parse()

	if mnemonic == "" {
		mnemonic, err = NewMnemonic()
		if err != nil {
			fmt.Printf("Error generating mnemonic: %v\n", err)
			return
		}
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		fmt.Println("mnemonic is invalid")
		return
	}

	fields := strings.Fields(mnemonic)
	if len(fields) != 24 {
		fmt.Println("mnemonic is invalid, must be 24 words")
		return
	}

	fmt.Println("mnemonic: ", mnemonic)
	fmt.Println("num words: ", len(fields))
	fmt.Println()

	{
		// cosmos
		privKey, err := NewAccount(mnemonic)
		if err != nil {
			fmt.Printf("Error generating account: %v\n", err)
			return
		}

		pubKey := privKey.PubKey()

		address := pubKey.Address()

		fmt.Println("cosmos private key: ", hex.EncodeToString(privKey.Bytes()))
		fmt.Println("cosmos public key: ", hex.EncodeToString(pubKey.Bytes()))

		fmt.Println("commos address: ", hex.EncodeToString(address.Bytes()))

		bech32, err := Bech32FromAddressBytes(address.Bytes())
		if err != nil {
			fmt.Printf("Error converting address to bech32: %v\n", err)
			return
		}

		fmt.Println("cosmos bech32 address: ", bech32)
		// testAddress(bech32)
	}

	fmt.Println()

	{
		// ethereum
		wallet, err := hdwallet.NewFromMnemonic(mnemonic)

		path := hdwallet.MustParseDerivationPath(defaultETHHDPath)
		account, err := wallet.Derive(path, false)
		if err != nil {
			fmt.Printf("Error deriving account: %v\n", err)
			return
		}

		privKey, err := wallet.PrivateKeyHex(account)
		if err != nil {
			fmt.Printf("Error getting private key: %v\n", err)
			return
		}
		pubKey, err := wallet.PublicKeyHex(account)
		if err != nil {
			fmt.Printf("Error getting public key: %v\n", err)
			return
		}

		address, err := wallet.AddressHex(account)
		if err != nil {
			fmt.Printf("Error getting address: %v\n", err)
			return
		}

		fmt.Println("eth private key: ", privKey)
		fmt.Println("eth public key: ", pubKey)
		fmt.Println("eth address: ", address)
	}

}
