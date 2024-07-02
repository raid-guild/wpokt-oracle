package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dan13ram/wpokt-oracle/common"
)

// Main Function
func main() {
	GoogleKeyName := os.Getenv("GCP_KMS_KEY_NAME")

	fmt.Println("Google KMS Key Name: ", GoogleKeyName)
	if GoogleKeyName == "" {
		log.Fatalf("GCP KMS Key Name not set")
	}

	signer, err := common.NewGcpKmsSigner(GoogleKeyName)
	if err != nil {
		log.Fatalf("failed to create GCP KMS signer: %v", err)
	}

	fmt.Println("Eth Address: ", signer.EthAddress())

	fmt.Println("Cosmos Public Key: ", signer.CosmosPublicKey())

	// Prepare the transaction data (example)
	txData := []byte("example transaction data")

	// Ethereum
	ethSignature, err := signer.EthSign(txData)
	if err != nil {
		log.Fatalf("failed to sign Ethereum hash: %v", err)
	}
	fmt.Printf("Ethereum Signature: %x\n", ethSignature)

	// Cosmos
	cosmosSignature, err := signer.CosmosSign(txData)
	if err != nil {
		log.Fatalf("failed to sign Cosmos hash: %v", err)
	}
	fmt.Printf("Cosmos Signature: %x\n", cosmosSignature)
}
