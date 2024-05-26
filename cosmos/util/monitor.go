package util

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/dan13ram/wpokt-oracle/app"
	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

func HexToBytes(hexStr string) ([]byte, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	return hex.DecodeString(hexStr)
}

func CreateTransaction(
	tx *sdk.TxResponse,
	chain models.Chain,
	senderAddress []byte,
	txStatus models.TxStatus,
) (models.Transaction, error) {

	txHash := strings.TrimPrefix(tx.TxHash, "0x")
	if len(txHash) != 64 {
		return models.Transaction{}, fmt.Errorf("invalid tx hash: %s", tx.TxHash)
	}

	senderAddressStr := strings.ToLower(hex.EncodeToString(senderAddress))
	if len(senderAddressStr) != 40 {
		return models.Transaction{}, fmt.Errorf("invalid sender address: %s", senderAddressStr)
	}

	return models.Transaction{
		Hash:        strings.ToLower(tx.TxHash),
		Sender:      senderAddressStr,
		BlockHeight: uint64(tx.Height),
		Chain:       chain,
		TxStatus:    txStatus,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func InsertTransaction(tx models.Transaction) error {
	err := app.DB.InsertOne(common.CollectionTransactions, tx)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil
		}
		return err
	}

	return nil
}

func UpdateTransaction(tx *models.Transaction, update bson.M) error {
	if tx == nil {
		return fmt.Errorf("tx is nil")
	}
	return app.DB.UpdateOne(
		common.CollectionTransactions,
		bson.M{"_id": tx.ID, "hash": tx.Hash},
		bson.M{"$set": update},
	)
}

func GetPendingTransactions() ([]models.Transaction, error) {
	txs := []models.Transaction{}

	err := app.DB.FindMany(common.CollectionTransactions, bson.M{"tx_status": models.TxStatusPending}, &txs)

	return txs, err
}

func ValidateMemo(txMemo string) (models.MintMemo, error) {
	var memo models.MintMemo

	err := json.Unmarshal([]byte(txMemo), &memo)
	if err != nil {
		return memo, fmt.Errorf("failed to unmarshal memo: %s", err)
	}

	memo.Address = strings.Trim(strings.ToLower(memo.Address), " ")
	memo.ChainID = strings.Trim(strings.ToLower(memo.ChainID), " ")

	if !common.IsValidEthereumAddress(memo.Address) {
		return memo, fmt.Errorf("invalid address: %s", memo.Address)
	}

	if strings.EqualFold(memo.Address, common.ZeroAddress) {
		return memo, fmt.Errorf("zero address: %s", memo.Address)
	}

	if !common.EthereumSupportedChainIDs[memo.ChainID] {
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
