package db

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

func NewMessageBody(
	senderAddress []byte,
	amount uint64,
	recipientAddress []byte,
) (models.MessageBody, error) {

	sender, err := common.AddressHexFromBytes(senderAddress)
	if err != nil {
		return models.MessageBody{}, err
	}

	recipient, err := common.AddressHexFromBytes(recipientAddress)
	if err != nil {
		return models.MessageBody{}, err
	}

	return models.MessageBody{
		SenderAddress:    sender,
		Amount:           amount,
		RecipientAddress: recipient,
	}, nil
}

func NewMessageContent(
	nonce uint32,
	originDomain uint32,
	senderAddress []byte,
	destinationDomain uint32,
	recipientAddress []byte,
	messageBody models.MessageBody,
) (models.MessageContent, error) {

	sender, err := common.AddressHexFromBytes(senderAddress)
	if err != nil {
		return models.MessageContent{}, err
	}

	recipient, err := common.AddressHexFromBytes(recipientAddress)
	if err != nil {
		return models.MessageContent{}, err
	}

	return models.MessageContent{
		Version:           common.HyperlaneVersion,
		Nonce:             nonce,
		OriginDomain:      originDomain,
		Sender:            sender,
		DestinationDomain: destinationDomain,
		Recipient:         recipient,
		MessageBody:       messageBody,
	}, nil
}

func NewMessage(
	originTxDoc *models.Transaction,
	content models.MessageContent,
	status models.MessageStatus,
) (models.Message, error) {
	messageIDBytes, err := content.MessageID()
	if err != nil {
		return models.Message{}, err
	}
	messageID := common.HexFromBytes(messageIDBytes)

	return models.Message{
		OriginTransaction:     originTxDoc.ID,
		OriginTransactionHash: originTxDoc.Hash,
		MessageID:             messageID,
		Content:               content,
		Signatures:            []models.Signature{},
		Transaction:           nil,
		Sequence:              nil,
		Status:                status,
		TransactionHash:       "",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}, nil
}

func NewMessageWithTxHash(
	originTxHash [32]byte,
	content models.MessageContent,
	status models.MessageStatus,
) (models.Message, error) {
	messageIDBytes, err := content.MessageID()
	if err != nil {
		return models.Message{}, err
	}
	messageID := common.HexFromBytes(messageIDBytes)

	txHash := common.HexFromBytes(originTxHash[:])

	return models.Message{
		OriginTransaction:     nil,
		OriginTransactionHash: txHash,
		MessageID:             messageID,
		Content:               content,
		Signatures:            []models.Signature{},
		Transaction:           nil,
		Sequence:              nil,
		Status:                status,
		TransactionHash:       "",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}, nil
}

func UpdateMessage(messageID *primitive.ObjectID, update bson.M) error {
	if messageID == nil {
		return fmt.Errorf("messageID is nil")
	}
	return mongoDB.UpdateOne(
		common.CollectionMessages,
		bson.M{"_id": messageID},
		bson.M{"$set": update},
	)
}

func InsertMessage(tx models.Message) (primitive.ObjectID, error) {
	insertedID, err := mongoDB.InsertOne(common.CollectionMessages, tx)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			var messageDoc models.Message
			if err = mongoDB.FindOne(common.CollectionMessages, bson.M{"origin_transaction_hash": tx.OriginTransactionHash}, &messageDoc); err != nil {
				return insertedID, err
			}
			return *messageDoc.ID, nil
		}
		return insertedID, err
	}

	return insertedID, nil
}
