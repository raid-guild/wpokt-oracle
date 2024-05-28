package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageContent struct {
	Version           uint8       `json:"version" bson:"version"`
	Nonce             uint32      `json:"nonce" bson:"nonce"`
	OriginDomain      uint32      `json:"origin_domain" bson:"origin_domain"`
	Sender            string      `json:"sender" bson:"sender"`
	DestinationDomain uint32      `json:"destination_domain" bson:"destination_domain"`
	Recipient         string      `json:"recipient" bson:"recipient"`
	MessageBody       MessageBody `json:"message_body" bson:"message_body"`
}

type Message struct {
	ID                    *primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	OriginTransaction     *primitive.ObjectID `json:"origin_transaction" bson:"origin_transaction"`
	OriginTransactionHash string              `json:"origin_transaction_hash" bson:"origin_transaction_hash"`
	MessageID             string              `json:"message_id" bson:"message_id"`
	Content               MessageContent      `json:"content" bson:"content"`
	Signatures            []Signature         `json:"signatures" bson:"signatures"`
	Transaction           primitive.ObjectID  `json:"transaction" bson:"transaction"`
	Sequence              uint64              `json:"sequence" bson:"sequence"` // account sequence for submitting the transaction
	TransactionHash       string              `json:"transaction_hash" bson:"transaction_hash"`
	CreatedAt             time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time           `bson:"updated_at" json:"updated_at"`
}

type MessageBody struct {
	SenderAddress    string `json:"sender_address" bson:"sender_address"`
	Amount           uint64 `json:"amount" bson:"amount"`
	RecipientAddress string `json:"recipient_address" bson:"recipient_address"`
}

type Signature struct {
	Signer    string `json:"signer" bson:"signer"`
	Signature string `json:"signature" bson:"signature"` // Assuming signature is a string representation
}
