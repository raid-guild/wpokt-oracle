package db

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

func NewEthereumTransaction(
	tx *types.Transaction,
	toAddress []byte,
	receipt *types.Receipt,
	chain models.Chain,
	txStatus models.TransactionStatus,
) (models.Transaction, error) {

	txHash := common.Ensure0xPrefix(receipt.TxHash.String())

	txTo, err := common.AddressHexFromBytes(toAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("invalid to address: %w", err)
	}

	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("could not get sender from tx: %w", err)
	}

	txFrom := common.Ensure0xPrefix(from.String())

	return models.Transaction{
		Hash:        txHash,
		FromAddress: txFrom,
		ToAddress:   txTo,
		BlockHeight: receipt.BlockNumber.Uint64(),
		Chain:       chain,
		Status:      txStatus,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Messages:    []primitive.ObjectID{},
	}, nil
}

func NewCosmosTransaction(
	txRes *sdk.TxResponse,
	chain models.Chain,
	fromAddress []byte,
	toAddress []byte,
	txStatus models.TransactionStatus,
) (models.Transaction, error) {

	txHash := common.Ensure0xPrefix(txRes.TxHash)
	if len(txHash) != 66 {
		return models.Transaction{}, fmt.Errorf("invalid tx hash")
	}

	txFrom, err := common.AddressHexFromBytes(fromAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("invalid from address: %w", err)
	}

	txTo, err := common.AddressHexFromBytes(toAddress)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("invalid to address: %w", err)
	}

	return models.Transaction{
		Hash:        txHash,
		FromAddress: txFrom,
		ToAddress:   txTo,
		BlockHeight: uint64(txRes.Height),
		Chain:       chain,
		Status:      txStatus,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Messages:    []primitive.ObjectID{},
	}, nil
}

func InsertTransaction(tx models.Transaction) (primitive.ObjectID, error) {
	insertedID, err := MongoDB.InsertOne(common.CollectionTransactions, tx)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			var txDoc models.Transaction
			if err = MongoDB.FindOne(common.CollectionTransactions, bson.M{"hash": tx.Hash}, &txDoc); err != nil {
				return insertedID, err
			}
			return *txDoc.ID, nil
		}
		return insertedID, err
	}

	return insertedID, nil
}

func UpdateTransaction(txID *primitive.ObjectID, update bson.M) error {
	if txID == nil {
		return fmt.Errorf("txID is nil")
	}
	_, err := MongoDB.UpdateOne(
		common.CollectionTransactions,
		bson.M{"_id": *txID},
		bson.M{"$set": update},
	)
	return err
}

func GetPendingTransactionsTo(chain models.Chain, toAddress []byte) ([]models.Transaction, error) {
	txs := []models.Transaction{}

	txTo, err := common.AddressHexFromBytes(toAddress)
	if err != nil {
		return txs, fmt.Errorf("invalid to address: %w", err)
	}

	filter := bson.M{
		"status":     models.TransactionStatusPending,
		"chain":      chain,
		"to_address": txTo,
	}

	err = MongoDB.FindMany(common.CollectionTransactions, filter, &txs)

	return txs, err
}

func GetConfirmedTransactionsTo(chain models.Chain, toAddress []byte) ([]models.Transaction, error) {
	txs := []models.Transaction{}

	txTo, err := common.AddressHexFromBytes(toAddress)
	if err != nil {
		return txs, fmt.Errorf("invalid to address: %w", err)
	}

	filter := bson.M{
		"status":     models.TransactionStatusConfirmed,
		"chain":      chain,
		"to_address": txTo,
	}

	refundNil := bson.M{
		"$or": []bson.M{
			{"refund": bson.M{"$exists": false}},
			{"refund": bson.M{"$eq": nil}},
		},
	}

	messagesEmpty := bson.M{
		"$or": []bson.M{
			{"messages": bson.M{"$exists": false}},
			{"messages": bson.M{"$eq": nil}},
			{"messages": bson.M{"$size": 0}},
		},
	}

	filter = bson.M{
		"$and": []bson.M{
			filter,
			refundNil,
			messagesEmpty,
		},
	}

	err = MongoDB.FindMany(common.CollectionTransactions, filter, &txs)

	return txs, err
}

func GetPendingTransactionsFrom(chain models.Chain, fromAddress []byte) ([]models.Transaction, error) {
	txs := []models.Transaction{}

	txFrom, err := common.AddressHexFromBytes(fromAddress)
	if err != nil {
		return txs, fmt.Errorf("invalid from address: %w", err)
	}

	filter := bson.M{
		"status":       models.TransactionStatusPending,
		"chain":        chain,
		"from_address": txFrom,
	}

	err = MongoDB.FindMany(common.CollectionTransactions, filter, &txs)

	return txs, err
}
