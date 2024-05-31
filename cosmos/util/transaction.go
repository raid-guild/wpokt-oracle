package util

import (
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/dan13ram/wpokt-oracle/app"
	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

func NewTransaction(
	tx *sdk.TxResponse,
	chain models.Chain,
	fromAddress []byte,
	toAddress []byte,
	txStatus models.TransactionStatus,
) (models.Transaction, error) {

	txHash := Ensure0xPrefix(tx.TxHash)
	if len(txHash) != 66 {
		return models.Transaction{}, fmt.Errorf("invalid tx hash: %s", tx.TxHash)
	}

	txFrom := Ensure0xPrefix(hex.EncodeToString(fromAddress))
	if len(txFrom) != 42 {
		return models.Transaction{}, fmt.Errorf("invalid from address: %s", txFrom)
	}

	txTo := Ensure0xPrefix(hex.EncodeToString(toAddress))
	if len(txTo) != 42 {
		return models.Transaction{}, fmt.Errorf("invalid to address: %s", txTo)
	}

	return models.Transaction{
		Hash:        txHash,
		FromAddress: txFrom,
		ToAddress:   txTo,
		BlockHeight: uint64(tx.Height),
		Chain:       chain,
		Status:      txStatus,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func InsertTransaction(tx models.Transaction) (primitive.ObjectID, error) {
	insertedID, err := app.DB.InsertOne(common.CollectionTransactions, tx)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			var refundDoc models.Transaction
			err := app.DB.FindOne(common.CollectionTransactions, bson.M{"hash": tx.Hash}, refundDoc)
			if err != nil {
				return insertedID, err
			}
			return *refundDoc.ID, nil
		}
		return insertedID, err
	}

	return insertedID, nil
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

func GetPendingTransactionsTo(chain models.Chain, toAddress []byte) ([]models.Transaction, error) {
	txs := []models.Transaction{}

	txTo := Ensure0xPrefix(hex.EncodeToString(toAddress))
	if len(txTo) != 42 {
		return txs, fmt.Errorf("invalid to address: %s", txTo)
	}

	filter := bson.M{
		"status":     models.TransactionStatusPending,
		"chain":      chain,
		"to_address": txTo,
	}

	err := app.DB.FindMany(common.CollectionTransactions, filter, &txs)

	return txs, err
}

func GetPendingTransactionsFrom(chain models.Chain, fromAddress []byte) ([]models.Transaction, error) {
	txs := []models.Transaction{}

	txFrom := Ensure0xPrefix(hex.EncodeToString(fromAddress))
	if len(txFrom) != 42 {
		return txs, fmt.Errorf("invalid from address: %s", txFrom)
	}

	filter := bson.M{
		"status":       models.TransactionStatusPending,
		"chain":        chain,
		"from_address": txFrom,
	}

	err := app.DB.FindMany(common.CollectionTransactions, filter, &txs)

	return txs, err
}
