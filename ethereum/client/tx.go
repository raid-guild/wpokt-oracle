package client

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

type ValidateTransactionByHashResult struct {
	Tx      *types.Transaction
	Receipt *types.Receipt
}

func ValidateTransactionByHash(client EthereumClient, txHash string) (*ValidateTransactionByHashResult, error) {
	tx, isPending, err := client.GetTransactionByHash(txHash)
	if err != nil {
		return nil, fmt.Errorf("error getting transaction by hash: %s", err)
	}
	if tx == nil {
		return nil, fmt.Errorf("transaction not found")
	}
	if isPending {
		return nil, fmt.Errorf("transaction is pending")
	}
	receipt, err := client.GetTransactionReceipt(txHash)
	if err != nil {
		return nil, fmt.Errorf("error getting transaction receipt: %s", err)
	}

	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("transaction failed")
	}
	return &ValidateTransactionByHashResult{tx, receipt}, nil
}
