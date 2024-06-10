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

func FindMessage(filter bson.M) (models.Message, error) {
	var message models.Message
	err := mongoDB.FindOne(common.CollectionMessages, filter, &message)
	return message, err
}

func UpdateMessage(messageID *primitive.ObjectID, update bson.M) error {
	if messageID == nil {
		return fmt.Errorf("messageID is nil")
	}
	_, err := mongoDB.UpdateOne(
		common.CollectionMessages,
		bson.M{"_id": messageID},
		bson.M{"$set": update},
	)
	return err
}

func UpdateMessageByMessageID(messageID [32]byte, update bson.M) (primitive.ObjectID, error) {
	messageIDHex := common.Ensure0xPrefix(common.HexFromBytes(messageID[:]))
	fmt.Println("messageIDHex", messageIDHex)
	return mongoDB.UpdateOne(
		common.CollectionMessages,
		bson.M{"message_id": messageIDHex},
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

func GetPendingMessages(signerToExclude string, chain models.Chain) ([]models.Message, error) {
	messages := []models.Message{}
	filter := bson.M{
		"$and": []bson.M{
			{"content.destination_domain": chain.ChainDomain},
			{"$or": []bson.M{
				{"status": models.MessageStatusPending},
				{"status": models.MessageStatusSigned},
			}},
			{"$nor": []bson.M{
				{"signatures": bson.M{
					"$elemMatch": bson.M{"signer": signerToExclude},
				}},
			}},
		},
	}

	err := mongoDB.FindMany(common.CollectionMessages, filter, &messages)

	return messages, err
}

func GetSignedMessages(chain models.Chain) ([]models.Message, error) {
	messages := []models.Message{}
	sort := bson.M{"sequence": 1}
	filter := bson.M{
		"content.destination_domain": chain.ChainDomain,
		"status":                     models.MessageStatusSigned,
	}

	err := mongoDB.FindManySorted(common.CollectionMessages, filter, sort, &messages)

	return messages, err
}

func GetBroadcastedMessages(chain models.Chain) ([]models.Message, error) {
	messages := []models.Message{}
	filter := bson.M{
		"content.destination_domain": chain.ChainDomain,
		"status":                     models.MessageStatusBroadcasted,
		"transaction":                nil,
	}

	err := mongoDB.FindMany(common.CollectionMessages, filter, &messages)

	return messages, err
}
