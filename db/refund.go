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

type RefundDB interface {
	NewRefund(
		txRes *sdk.TxResponse,
		txDoc *models.Transaction,
		recipientAddress []byte,
		amountCoin sdk.Coin,
	) (models.Refund, error)

	InsertRefund(tx models.Refund) (primitive.ObjectID, error)
	UpdateRefund(refundID *primitive.ObjectID, update bson.M) error

	GetPendingRefunds(signerToExclude string) ([]models.Refund, error)
	GetSignedRefunds() ([]models.Refund, error)
	GetBroadcastedRefunds() ([]models.Refund, error)
}

func newRefund(
	txRes *sdk.TxResponse,
	txDoc *models.Transaction,
	recipientAddress []byte,
	amountCoin sdk.Coin,
) (models.Refund, error) {

	if txRes == nil || txDoc == nil || txDoc.Hash == "" || txDoc.ID == nil || txRes.TxHash == "" {
		return models.Refund{}, fmt.Errorf("txRes or txDoc is nil")
	}

	txHash := common.Ensure0xPrefix(txRes.TxHash)
	if txHash != txDoc.Hash {
		return models.Refund{}, fmt.Errorf("tx hash mismatch")
	}

	recipient, err := common.AddressHexFromBytes(recipientAddress)
	if err != nil {
		return models.Refund{}, fmt.Errorf("invalid recipient address: %w", err)
	}

	amount := amountCoin.Amount.String()

	return models.Refund{
		OriginTransaction:     *txDoc.ID,
		OriginTransactionHash: txDoc.Hash,
		Recipient:             recipient,
		Amount:                amount,
		Signatures:            []models.Signature{},
		Transaction:           nil,
		Sequence:              nil,
		Status:                models.RefundStatusPending,
		TransactionHash:       "",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}, nil
}

func insertRefund(tx models.Refund) (primitive.ObjectID, error) {
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

func updateRefund(refundID *primitive.ObjectID, update bson.M) error {
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

func getPendingRefunds(signerToExclude string) ([]models.Refund, error) {
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

func getSignedRefunds() ([]models.Refund, error) {
	refunds := []models.Refund{}
	filter := bson.M{"status": models.RefundStatusSigned}
	sort := bson.M{"sequence": 1}

	err := mongoDB.FindManySorted(common.CollectionRefunds, filter, sort, &refunds)

	return refunds, err
}

func getBroadcastedRefunds() ([]models.Refund, error) {
	refunds := []models.Refund{}
	filter := bson.M{"status": models.RefundStatusBroadcasted, "transaction": nil}

	err := mongoDB.FindMany(common.CollectionRefunds, filter, &refunds)

	return refunds, err
}

type refundDB struct{}

func (db *refundDB) NewRefund(
	txRes *sdk.TxResponse,
	txDoc *models.Transaction,
	recipientAddress []byte,
	amountCoin sdk.Coin,
) (models.Refund, error) {
	return newRefund(txRes, txDoc, recipientAddress, amountCoin)
}

func (db *refundDB) InsertRefund(tx models.Refund) (primitive.ObjectID, error) {
	return insertRefund(tx)
}

func (db *refundDB) UpdateRefund(refundID *primitive.ObjectID, update bson.M) error {
	return updateRefund(refundID, update)
}

func (db *refundDB) GetPendingRefunds(signerToExclude string) ([]models.Refund, error) {
	return getPendingRefunds(signerToExclude)
}

func (db *refundDB) GetSignedRefunds() ([]models.Refund, error) {
	return getSignedRefunds()
}

func (db *refundDB) GetBroadcastedRefunds() ([]models.Refund, error) {
	return getBroadcastedRefunds()
}
