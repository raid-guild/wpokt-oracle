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

func NewRefund(
	txRes *sdk.TxResponse,
	txDoc *models.Transaction,
	recipientAddress []byte,
	amountCoin sdk.Coin,
	txBody string,
) (models.Refund, error) {

	if txRes == nil || txDoc == nil {
		return models.Refund{}, fmt.Errorf("txRes or txDoc is nil")
	}

	txHash := common.Ensure0xPrefix(txRes.TxHash)
	if txHash != txDoc.Hash {
		return models.Refund{}, fmt.Errorf("tx hash mismatch: %s != %s", txHash, txDoc.Hash)
	}

	recipient, err := common.AddressHexFromBytes(recipientAddress)
	if err != nil {
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
		Sequence:              nil,
		Status:                models.RefundStatusPending,
		TransactionHash:       "",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}, nil
}

func InsertRefund(tx models.Refund) (primitive.ObjectID, error) {
	insertedID, err := mongoDB.InsertOne(common.CollectionRefunds, tx)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			var refundDoc models.Refund
			if err = mongoDB.FindOne(common.CollectionRefunds, bson.M{"origin_transaction_hash": tx.OriginTransactionHash}, &refundDoc); err != nil {
				return insertedID, err
			}
			return *refundDoc.ID, nil
		}
		return insertedID, err
	}

	return insertedID, nil
}

func UpdateRefund(refundID *primitive.ObjectID, update bson.M) error {
	if refundID == nil {
		return fmt.Errorf("refundID is nil")
	}
	_, err := mongoDB.UpdateOne(
		common.CollectionRefunds,
		bson.M{"_id": refundID},
		bson.M{"$set": update},
	)
	return err
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

	err := mongoDB.FindMany(common.CollectionRefunds, filter, &refunds)

	return refunds, err
}

func GetSignedRefunds() ([]models.Refund, error) {
	refunds := []models.Refund{}
	filter := bson.M{"status": models.RefundStatusSigned}
	sort := bson.M{"sequence": 1}

	err := mongoDB.FindManySorted(common.CollectionRefunds, filter, sort, &refunds)

	return refunds, err
}

func GetBroadcastedRefunds() ([]models.Refund, error) {
	refunds := []models.Refund{}
	filter := bson.M{"status": models.RefundStatusBroadcasted, "transaction": nil}

	err := mongoDB.FindMany(common.CollectionRefunds, filter, &refunds)

	return refunds, err
}
