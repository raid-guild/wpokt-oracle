package util

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/dan13ram/wpokt-oracle/app"
	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

func HexToBytes(hexStr string) ([]byte, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	return hex.DecodeString(hexStr)
}

func Ensure0xPrefix(str string) string {
	str = strings.ToLower(str)
	if !strings.HasPrefix(str, "0x") {
		return "0x" + str
	}
	return str
}

func HexFromBytes(address []byte) string {
	return Ensure0xPrefix(hex.EncodeToString(address))
}

func CreateRefund(
	txRes *sdk.TxResponse,
	txDoc *models.Transaction,
	recipientAddress []byte,
	amountCoin sdk.Coin,
	txBody string,
) (models.Refund, error) {

	if txRes == nil || txDoc == nil {
		return models.Refund{}, fmt.Errorf("txRes or txDoc is nil")
	}

	txHash := Ensure0xPrefix(txRes.TxHash)
	if txHash != txDoc.Hash {
		return models.Refund{}, fmt.Errorf("tx hash mismatch: %s != %s", txHash, txDoc.Hash)
	}

	recipient := Ensure0xPrefix(hex.EncodeToString(recipientAddress))
	if len(recipient) != 42 {
		return models.Refund{}, fmt.Errorf("invalid recipient address: %s", recipient)
	}

	amount := amountCoin.Amount.Uint64()

	return models.Refund{
		OriginTransaction:     txDoc.ID,
		OriginTransactionHash: txDoc.Hash,
		Recipient:             recipient,
		Amount:                amount,
		TransactionBody:       txBody,
		Signatures:            []models.Signature{},
		Transaction:           nil,
		Status:                models.RefundStatusPending,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}, nil
}

func InsertRefund(tx models.Refund) (primitive.ObjectID, error) {
	insertedID, err := app.DB.InsertOne(common.CollectionRefunds, tx)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			var refundDoc models.Refund
			err := app.DB.FindOne(common.CollectionRefunds, bson.M{"origin_transaction_hash": tx.OriginTransactionHash}, refundDoc)
			if err != nil {
				return insertedID, err
			}
			return *refundDoc.ID, nil
		}
		return insertedID, err
	}

	return insertedID, nil
}

type ResultMaxSequence struct {
	MaxSequence int `bson:"max_sequence"`
}

func FindMaxSequenceFromRefunds() (uint64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "max_sequence", Value: bson.D{{Key: "$max", Value: "$sequence"}}},
		}}},
	}

	var result ResultMaxSequence
	err := app.DB.AggregateOne(common.CollectionRefunds, pipeline, &result)
	if err != nil {
		return 0, err
	}

	return uint64(result.MaxSequence), nil
}

func FindMaxSequenceFromMessages(chain models.Chain) (uint64, error) {
	filter := bson.M{"chain": chain}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "max_sequence", Value: bson.D{{Key: "$max", Value: "$sequence"}}},
		}}},
	}

	var result ResultMaxSequence
	err := app.DB.AggregateOne(common.CollectionMessages, pipeline, &result)
	if err != nil {
		return 0, err
	}

	return uint64(result.MaxSequence), nil
}

func FindMaxSequence(chain models.Chain) (uint64, error) {
	maxSequenceRefunds, err := FindMaxSequenceFromRefunds()
	if err != nil {
		return 0, err
	}

	maxSequenceMessages, err := FindMaxSequenceFromMessages(chain)
	if err != nil {
		return 0, err
	}

	if maxSequenceRefunds > maxSequenceMessages {
		return maxSequenceRefunds, nil
	}

	return maxSequenceMessages, nil
}

func UpdateMessage(messageID *primitive.ObjectID, update bson.M) error {
	if messageID == nil {
		return fmt.Errorf("messageID is nil")
	}
	return app.DB.UpdateOne(
		common.CollectionMessages,
		bson.M{"_id": messageID},
		bson.M{"$set": update},
	)
}

func UpdateRefund(refundID *primitive.ObjectID, update bson.M) error {
	if refundID == nil {
		return fmt.Errorf("refundID is nil")
	}
	return app.DB.UpdateOne(
		common.CollectionRefunds,
		bson.M{"_id": refundID},
		bson.M{"$set": update},
	)
}

func CreateTransaction(
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

func GetPendingRefunds(signerToExclude string) ([]models.Refund, error) {
	refunds := []models.Refund{}
	filter := bson.M{
		"$and": []bson.M{
			{"$or": []bson.M{
				{"status": models.RefundStatusPending},
				{"status": models.RefundStatusSigned},
			}},
			{"$nor": []bson.M{
				{"signatures": bson.M{
					"$elemMatch": bson.M{"signer": signerToExclude},
				}},
			}},
		},
	}

	err := app.DB.FindMany(common.CollectionRefunds, filter, &refunds)

	return refunds, err
}

func GetSignedRefunds() ([]models.Refund, error) {
	refunds := []models.Refund{}
	filter := bson.M{"status": models.RefundStatusSigned}
	sort := bson.M{"sequence": 1}

	err := app.DB.FindManySorted(common.CollectionRefunds, filter, sort, &refunds)

	return refunds, err
}

func GetBroadcastedRefunds() ([]models.Refund, error) {
	refunds := []models.Refund{}
	filter := bson.M{"status": models.RefundStatusBroadcasted}

	err := app.DB.FindMany(common.CollectionRefunds, filter, &refunds)

	return refunds, err
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
