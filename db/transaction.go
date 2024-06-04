package db

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

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

	txHash := common.Ensure0xPrefix(tx.TxHash)
	if len(txHash) != 66 {
		return models.Transaction{}, fmt.Errorf("invalid tx hash: %s", tx.TxHash)
	}

	txFrom, err := common.AddressHexFromBytes(fromAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("invalid from address: %s", txFrom)
	}

	txTo, err := common.AddressHexFromBytes(toAddress)
	if err != nil {
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
	insertedID, err := mongoDB.InsertOne(common.CollectionTransactions, tx)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			var txDoc models.Transaction
			if err = mongoDB.FindOne(common.CollectionTransactions, bson.M{"hash": tx.Hash}, &txDoc); err != nil {
				return insertedID, err
			}
			return *txDoc.ID, nil
		}
		return insertedID, err
	}

	return insertedID, nil
}

func UpdateTransaction(tx *models.Transaction, update bson.M) error {
	if tx == nil {
		return fmt.Errorf("tx is nil")
	}
	return mongoDB.UpdateOne(
		common.CollectionTransactions,
		bson.M{"_id": tx.ID, "hash": tx.Hash},
		bson.M{"$set": update},
	)
}

func GetPendingTransactionsTo(chain models.Chain, toAddress []byte) ([]models.Transaction, error) {
	txs := []models.Transaction{}

	txTo, err := common.AddressHexFromBytes(toAddress)
	if err != nil {
		return txs, fmt.Errorf("invalid to address: %s", txTo)
	}

	filter := bson.M{
		"status":     models.TransactionStatusPending,
		"chain":      chain,
		"to_address": txTo,
	}

	err = mongoDB.FindMany(common.CollectionTransactions, filter, &txs)

	return txs, err
}

func GetConfirmedTransactionsTo(chain models.Chain, toAddress []byte) ([]models.Transaction, error) {
	txs := []models.Transaction{}

	txTo, err := common.AddressHexFromBytes(toAddress)
	if err != nil {
		return txs, fmt.Errorf("invalid to address: %s", txTo)
	}

	filter := bson.M{
		"status":     models.TransactionStatusConfirmed,
		"chain":      chain,
		"to_address": txTo,
	}

	err = mongoDB.FindMany(common.CollectionTransactions, filter, &txs)

	return txs, err
}

func GetPendingTransactionsFrom(chain models.Chain, fromAddress []byte) ([]models.Transaction, error) {
	txs := []models.Transaction{}

	txFrom, err := common.AddressHexFromBytes(fromAddress)
	if err != nil {
		return txs, fmt.Errorf("invalid from address: %s", txFrom)
	}

	filter := bson.M{
		"status":       models.TransactionStatusPending,
		"chain":        chain,
		"from_address": txFrom,
	}

	err = mongoDB.FindMany(common.CollectionTransactions, filter, &txs)

	return txs, err
}
