package db

import (
	"fmt"
	"math/big"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

type MessageDB interface {
	NewMessageBody(
		senderAddress []byte,
		amount *big.Int,
		recipientAddress []byte,
	) (models.MessageBody, error)

	NewMessageContent(
		nonce uint32,
		originDomain uint32,
		senderAddress []byte,
		destinationDomain uint32,
		recipientAddress []byte,
		messageBody models.MessageBody,
	) (models.MessageContent, error)

	NewMessage(
		txDoc *models.Transaction,
		content models.MessageContent,
		status models.MessageStatus,
	) (models.Message, error)

	FindMessage(filter bson.M) (models.Message, error)

	UpdateMessage(messageID *primitive.ObjectID, update bson.M) error

	UpdateMessageByMessageID(messageID [32]byte, update bson.M) (primitive.ObjectID, error)

	InsertMessage(tx models.Message) (primitive.ObjectID, error)

	GetPendingMessages(signerToExclude string, chain models.Chain) ([]models.Message, error)

	GetSignedMessages(chain models.Chain) ([]models.Message, error)

	GetBroadcastedMessages(chain models.Chain) ([]models.Message, error)
}

func newMessageBody(
	senderAddress []byte,
	amount *big.Int,
	recipientAddress []byte,
) (models.MessageBody, error) {

	sender, err := common.AddressHexFromBytes(senderAddress)
	if err != nil {
		return models.MessageBody{}, fmt.Errorf("invalid sender address: %w", err)
	}

	recipient, err := common.AddressHexFromBytes(recipientAddress)
	if err != nil {
		return models.MessageBody{}, fmt.Errorf("invalid recipient address: %w", err)
	}

	return models.MessageBody{
		SenderAddress:    sender,
		Amount:           amount.String(),
		RecipientAddress: recipient,
	}, nil
}

func newMessageContent(
	nonce uint32,
	originDomain uint32,
	senderAddress []byte,
	destinationDomain uint32,
	recipientAddress []byte,
	messageBody models.MessageBody,
) (models.MessageContent, error) {

	sender, err := common.AddressHexFromBytes(senderAddress)
	if err != nil {
		return models.MessageContent{}, fmt.Errorf("invalid sender address: %w", err)
	}

	recipient, err := common.AddressHexFromBytes(recipientAddress)
	if err != nil {
		return models.MessageContent{}, fmt.Errorf("invalid recipient address: %w", err)
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

func newMessage(
	txDoc *models.Transaction,
	content models.MessageContent,
	status models.MessageStatus,
) (models.Message, error) {
	if (txDoc == nil) || (txDoc.ID == nil) || (txDoc.Hash == "") {
		return models.Message{}, fmt.Errorf("invalid txDoc")
	}

	messageIDBytes, err := content.MessageID()
	if err != nil {
		return models.Message{}, err
	}
	messageID := common.HexFromBytes(messageIDBytes)

	return models.Message{
		OriginTransaction:     *txDoc.ID,
		OriginTransactionHash: txDoc.Hash,
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

func findMessage(filter bson.M) (models.Message, error) {
	var message models.Message
	err := mongoDB.FindOne(common.CollectionMessages, filter, &message)
	return message, err
}

func updateMessage(messageID *primitive.ObjectID, update bson.M) error {
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

func updateMessageByMessageID(messageID [32]byte, update bson.M) (primitive.ObjectID, error) {
	messageIDHex := common.Ensure0xPrefix(common.HexFromBytes(messageID[:]))

	return mongoDB.UpdateOne(
		common.CollectionMessages,
		bson.M{"message_id": messageIDHex},
		bson.M{"$set": update},
	)
}

func insertMessage(tx models.Message) (primitive.ObjectID, error) {
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

func getPendingMessages(signerToExclude string, chain models.Chain) ([]models.Message, error) {
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
	sort := bson.M{"content.nonce": 1}

	err := mongoDB.FindManySorted(common.CollectionMessages, filter, sort, &messages)

	return messages, err
}

func getSignedMessages(chain models.Chain) ([]models.Message, error) {
	messages := []models.Message{}
	sort := bson.M{"sequence": 1}
	filter := bson.M{
		"content.destination_domain": chain.ChainDomain,
		"status":                     models.MessageStatusSigned,
	}

	err := mongoDB.FindManySorted(common.CollectionMessages, filter, sort, &messages)

	return messages, err
}

func getBroadcastedMessages(chain models.Chain) ([]models.Message, error) {
	messages := []models.Message{}
	filter := bson.M{
		"content.destination_domain": chain.ChainDomain,
		"status":                     models.MessageStatusBroadcasted,
		"transaction":                nil,
	}

	err := mongoDB.FindMany(common.CollectionMessages, filter, &messages)

	return messages, err
}

type messageDB struct{}

func (db *messageDB) NewMessageBody(
	senderAddress []byte,
	amount *big.Int,
	recipientAddress []byte,
) (models.MessageBody, error) {
	return newMessageBody(senderAddress, amount, recipientAddress)
}

func (db *messageDB) NewMessageContent(
	nonce uint32,
	originDomain uint32,
	senderAddress []byte,
	destinationDomain uint32,
	recipientAddress []byte,
	messageBody models.MessageBody,
) (models.MessageContent, error) {
	return newMessageContent(nonce, originDomain, senderAddress, destinationDomain, recipientAddress, messageBody)
}

func (db *messageDB) NewMessage(
	txDoc *models.Transaction,
	content models.MessageContent,
	status models.MessageStatus,
) (models.Message, error) {
	return newMessage(txDoc, content, status)
}

func (db *messageDB) FindMessage(filter bson.M) (models.Message, error) {
	return findMessage(filter)
}

func (db *messageDB) UpdateMessage(messageID *primitive.ObjectID, update bson.M) error {
	return updateMessage(messageID, update)
}

func (db *messageDB) UpdateMessageByMessageID(messageID [32]byte, update bson.M) (primitive.ObjectID, error) {
	return updateMessageByMessageID(messageID, update)
}

func (db *messageDB) InsertMessage(tx models.Message) (primitive.ObjectID, error) {
	return insertMessage(tx)
}

func (db *messageDB) GetPendingMessages(signerToExclude string, chain models.Chain) ([]models.Message, error) {
	return getPendingMessages(signerToExclude, chain)
}

func (db *messageDB) GetSignedMessages(chain models.Chain) ([]models.Message, error) {
	return getSignedMessages(chain)
}

func (db *messageDB) GetBroadcastedMessages(chain models.Chain) ([]models.Message, error) {
	return getBroadcastedMessages(chain)
}
