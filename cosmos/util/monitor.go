package util

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/dan13ram/wpokt-oracle/models"
)

func CreateTransaction(tx *sdk.TxResponse, senderAddress []byte) (models.Transaction, error) {
	hash0x := fmt.Sprintf("0x%s", strings.ToLower(tx.TxHash))
	hashBytes, err := hex.DecodeString(hash0x)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to decode tx hash: %s", err)
	}

	return models.Transaction{
		Hash:               hashBytes,
		Sender:             senderAddress,
		BlockHeight:        uint64(tx.Height),
		BlockConfirmations: 0,
		// 	ChainType:
		// ChainID
		// ChainDomain
		TxStatus: "failed",
		IsValid:  false,
		Refund: models.RefundInfo{
			Required:     false,
			Refunded:     false,
			RefundTxHash: []byte{},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

var SupportedChainIDs = map[string]bool{
	"1":        true,
	"11155111": true,
	"31337":    true,
}

func ValidateMemo(txMemo string) (models.MintMemo, error) {
	var memo models.MintMemo

	err := json.Unmarshal([]byte(txMemo), &memo)
	if err != nil {
		return memo, fmt.Errorf("failed to unmarshal memo: %s", err)
	}

	memo.Address = strings.Trim(strings.ToLower(memo.Address), " ")
	memo.ChainID = strings.Trim(strings.ToLower(memo.ChainID), " ")

	address := common.HexToAddress(memo.Address).Hex()
	if !strings.EqualFold(address, memo.Address) {
		return memo, fmt.Errorf("invalid address: %s", memo.Address)
	}

	if address == common.HexToAddress("").Hex() {
		return memo, fmt.Errorf("zero address: %s", memo.Address)
	}

	if !SupportedChainIDs[memo.ChainID] {
		return memo, fmt.Errorf("unsupported chain id: %s", memo.ChainID)
	}

	return memo, nil
}

/*
// func CreateMint(sdk *pokt.TxResponse, memo models.MintMemo, wpoktAddress string, vaultAddress string) models.Mint {
// 	return models.Mint{
// 		Height:          strconv.FormatInt(tx.Height, 10),
// 		Confirmations:   "0",
// 		TransactionHash: strings.ToLower(tx.Hash),
// 		SenderAddress:   strings.ToLower(tx.StdTx.Msg.Value.FromAddress),
// 		// SenderChainID:       app.Config.Pocket.ChainID,
// 		RecipientAddress:    strings.ToLower(memo.Address),
// 		RecipientChainID:    memo.ChainID,
// 		WPOKTAddress:        strings.ToLower(wpoktAddress),
// 		VaultAddress:        strings.ToLower(vaultAddress),
// 		Amount:              tx.StdTx.Msg.Value.Amount,
// 		Memo:                &memo,
// 		CreatedAt:           time.Now(),
// 		UpdatedAt:           time.Now(),
// 		Status:              models.StatusPending,
// 		Data:                nil,
// 		MintTransactionHash: "",
// 		Signers:             []string{},
// 		Signatures:          []string{},
// 	}
// }
//
// func CreateInvalidMint(tx *pokt.TxResponse, vaultAddress string) models.InvalidMint {
// 	return models.InvalidMint{
// 		Height:          strconv.FormatInt(tx.Height, 10),
// 		Confirmations:   "0",
// 		TransactionHash: strings.ToLower(tx.Hash),
// 		SenderAddress:   strings.ToLower(tx.StdTx.Msg.Value.FromAddress),
// 		// SenderChainID:   app.Config.Pocket.ChainID,
// 		Memo:         tx.StdTx.Memo,
// 		Amount:       tx.StdTx.Msg.Value.Amount,
// 		VaultAddress: strings.ToLower(vaultAddress),
// 		CreatedAt:    time.Now(),
// 		UpdatedAt:    time.Now(),
// 		Status:       models.StatusPending,
// 		Signers:      []string{},
// 		ReturnTx:     "",
// 		ReturnTxHash: "",
// 	}
// }
*/
