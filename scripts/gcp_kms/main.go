package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"

	"github.com/dan13ram/wpokt-oracle/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

// Main Function
func main() {
	GoogleAppCredsFilePath := os.Getenv("GCP_CREDS_JSON")
	GoogleKeyName := os.Getenv("GCP_KMS_KEY_NAME")

	fmt.Println("Google App Creds Path: ", GoogleAppCredsFilePath)
	fmt.Println("Google KMS Key Name: ", GoogleKeyName)

	signer, err := common.NewGcpKmsSigner(GoogleAppCredsFilePath, GoogleKeyName)
	if err != nil {
		log.Fatalf("failed to create GCP KMS signer: %v", err)
	}

	fmt.Println("Eth Address: ", signer.EthAddress())

	fmt.Println("Cosmos Public Key: ", signer.CosmosPublicKey())

	// Prepare the transaction data (example)
	txData := []byte("example transaction data")
	hash := sha256.Sum256(txData)

	// Ethereum
	ethSignature, err := signer.EthSignHash(ethcommon.BytesToHash(hash[:]))
	if err != nil {
		log.Fatalf("failed to sign Ethereum hash: %v", err)
	}
	fmt.Printf("Ethereum Signature: %x\n", ethSignature)

	// Cosmos
	cosmosSignature, err := signer.CosmosSignHash(hash)
	if err != nil {
		log.Fatalf("failed to sign Cosmos hash: %v", err)
	}
	fmt.Printf("Cosmos Signature: %x\n", cosmosSignature)
}
