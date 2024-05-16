package main

import (
	"encoding/hex"
	"flag"
	"log"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/dan13ram/wpokt-oracle/common"

	"github.com/dan13ram/wpokt-oracle/scripts/utils"
)

func main() {
	var publickeys string
	var threshold int
	flag.StringVar(&publickeys, "publickeys", "", "comma separated list of public keys")
	flag.IntVar(&threshold, "threshold", 0, "threshold for multisig")
	flag.Parse()

	if publickeys == "" {
		log.Fatal("publickeys is required")
	}

	var keys []string
	if publickeys != "" {
		keys = strings.Split(publickeys, ",")
	}

	if threshold <= 0 {
		log.Fatal("threshold is required")
	}

	if len(keys) < threshold {
		log.Fatal("threshold must be less than or equal to the number of public keys")
	}

	var pKeys []crypto.PubKey

	for i, key := range keys {
		if !common.IsValidCosmosPublicKey(key) {
			log.Fatalf("invalid public key %d: %v", i, key)
		}
		pKey := &secp256k1.PubKey{}
		pKeyBytes, err := hex.DecodeString(key)
		if err != nil {
			log.Fatalf("error decoding public key %d: %v", i, err)
		}
		err = pKey.UnmarshalAmino(pKeyBytes)
		if err != nil {
			log.Fatalf("error unmarshalling public key %d: %v", i, err)
		}
		pKeys = append(pKeys, pKey)
		log.Printf("public key %d: %v", i, key)
	}

	log.Printf("threshold: %v", threshold)
	pk := multisig.NewLegacyAminoPubKey(threshold, pKeys)
	log.Printf("multisig address: %v", strings.ToLower(pk.Address().String()))

	bech32, err := utils.Bech32FromAddressBytes(pk.Address().Bytes())
	if err != nil {
		log.Fatalf("error encoding address: %v", err)
	}
	log.Printf("multisig address bech32: %v", bech32)

	// utils.TestAddress(bech32)
}
