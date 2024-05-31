package util

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dan13ram/wpokt-oracle/app"
	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

func NewMessageBody(
	senderAddress string,
	amount uint64,
	recipientAddress string,
) models.MessageBody {
	return models.MessageBody{
		SenderAddress:    senderAddress,
		Amount:           amount,
		RecipientAddress: recipientAddress,
	}
}

func NewMessageContent(
	version uint8,
	nonce uint32,
	originDomain uint32,
	sender string,
	destinationDomain uint32,
	recipient string,
	messageBody models.MessageBody,
) models.MessageContent {
	return models.MessageContent{
		Version:           version,
		Nonce:             nonce,
		OriginDomain:      originDomain,
		Sender:            sender,
		DestinationDomain: destinationDomain,
		Recipient:         recipient,
		MessageBody:       messageBody,
	}
}

func NewMessage(
	originTransaction *primitive.ObjectID,
	originTransactionHash string,
	messageID string,
	content models.MessageContent,
	signatures []models.Signature,
	transaction primitive.ObjectID,
	sequence uint64,
	status models.MessageStatus,
	transactionHash string,
) *models.Message {
	return &models.Message{
		OriginTransaction:     originTransaction,
		OriginTransactionHash: originTransactionHash,
		MessageID:             messageID,
		Content:               content,
		Signatures:            signatures,
		Transaction:           transaction,
		Sequence:              sequence,
		Status:                status,
		TransactionHash:       transactionHash,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
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
