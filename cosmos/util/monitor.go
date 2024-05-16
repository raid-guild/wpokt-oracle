package util

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

/*
// func CreateMint(sdk *pokt.TxResponse, memo models.MintMemo, wpoktAddress string, vaultAddress string) models.Mint {
// 	return models.Mint{
// 		Height:          strconv.FormatInt(tx.Height, 10),
// 		Confirmations:   "0",
// 		TransactionHash: strings.ToLower(tx.Hash),
// 		SenderAddress:   strings.ToLower(tx.StdTx.Msg.Value.FromAddress),
// 		// SenderChainId:       app.Config.Pocket.ChainId,
// 		RecipientAddress:    strings.ToLower(memo.Address),
// 		RecipientChainId:    memo.ChainId,
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
// func ValidateMemo(txMemo string) (models.MintMemo, bool) {
// 	var memo models.MintMemo
//
// 	err := json.Unmarshal([]byte(txMemo), &memo)
// 	if err != nil {
// 		return memo, false
// 	}
//
// 	address := common.HexToAddress(memo.Address).Hex()
// 	if !strings.EqualFold(address, memo.Address) {
// 		return memo, false
// 	}
//
// 	if address == common.HexToAddress("").Hex() {
// 		return memo, false
// 	}
// 	memo.Address = address
//
// 	memoChainId, err := strconv.Atoi(memo.ChainId)
// 	if err != nil {
// 		return memo, false
// 	}
//
// 	appChainId, err := strconv.Atoi("1")
// 	if err != nil {
// 		return memo, false
// 	}
//
// 	if memoChainId != appChainId {
// 		return memo, false
// 	}
// 	memo.ChainId = "1"
// 	return memo, true
// }
// func CreateInvalidMint(tx *pokt.TxResponse, vaultAddress string) models.InvalidMint {
// 	return models.InvalidMint{
// 		Height:          strconv.FormatInt(tx.Height, 10),
// 		Confirmations:   "0",
// 		TransactionHash: strings.ToLower(tx.Hash),
// 		SenderAddress:   strings.ToLower(tx.StdTx.Msg.Value.FromAddress),
// 		// SenderChainId:   app.Config.Pocket.ChainId,
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
