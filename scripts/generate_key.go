package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"strings"

	"github.com/cosmos/go-bip39"

	hdwallet "github.com/dan13ram/go-ethereum-hdwallet"

	"github.com/dan13ram/wpokt-oracle/scripts/utils"
)

func main() {

	var mnemonic string
	var insecure bool
	var err error
	flag.StringVar(&mnemonic, "mnemonic", "", "24 word mnemonic")
	flag.BoolVar(&insecure, "insecure", false, "insecure mode")
	flag.Parse()

	if mnemonic == "" {
		mnemonic, err = utils.NewMnemonic()
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
	if !insecure && len(fields) != 24 {
		fmt.Println("mnemonic is invalid, must be 24 words")
		return
	}

	fmt.Println("mnemonic: ", mnemonic)
	fmt.Println("num words: ", len(fields))
	fmt.Println()

	{
		// cosmos
		privKey, err := utils.NewAccount(mnemonic)
		if err != nil {
			fmt.Printf("Error generating account: %v\n", err)
			return
		}

		pubKey := privKey.PubKey()

		address := pubKey.Address()

		fmt.Println("cosmos private key: ", hex.EncodeToString(privKey.Bytes()))
		fmt.Println("cosmos public key: ", hex.EncodeToString(pubKey.Bytes()))

		fmt.Println("commos address: ", hex.EncodeToString(address.Bytes()))

		bech32, err := utils.Bech32FromAddressBytes(address.Bytes())
		if err != nil {
			fmt.Printf("Error converting address to bech32: %v\n", err)
			return
		}

		fmt.Println("cosmos bech32 address: ", bech32)

		// utils.TestAddress(bech32)
	}

	fmt.Println()

	{
		// ethereum
		wallet, err := hdwallet.NewFromMnemonic(mnemonic)

		path := hdwallet.MustParseDerivationPath(utils.DefaultETHHDPath)
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
