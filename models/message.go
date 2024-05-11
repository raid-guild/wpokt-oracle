package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID                     primitive.ObjectID `json:"id" bson:"_id"`
	Version                uint8              `json:"version" bson:"version"`
	Nonce                  uint32             `json:"nonce" bson:"nonce"`
	OriginDomain           uint32             `json:"originDomain" bson:"originDomain"`
	Sender                 []byte             `json:"sender" bson:"sender"`
	DestinationDomain      uint32             `json:"destinationDomain" bson:"destinationDomain"`
	Recipient              []byte             `json:"recipient" bson:"recipient"`
	MessageBody            MessageBody        `json:"messageBody" bson:"messageBody"`
	Signatures             []Signature        `json:"signatures" bson:"signatures"`
	SignatureThreshold     int                `json:"signatureThreshold" bson:"signatureThreshold"`
	OriginTransaction      primitive.ObjectID `json:"originTransaction" bson:"originTransaction"`
	DestinationTransaction primitive.ObjectID `json:"destinationTransaction" bson:"destinationTransaction"`
	CreatedAt              time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt              time.Time          `bson:"updated_at" json:"updated_at"`
	UpdatedBy              primitive.ObjectID `bson:"updated_by" json:"updated_by"`
}

type MessageBody struct {
	SenderAddress    []byte `json:"senderAddress" bson:"senderAddress"`
	Amount           uint64 `json:"amount" bson:"amount"`
	RecipientAddress []byte `json:"recipientAddress" bson:"recipientAddress"`
}

type Signature struct {
	Signer    []byte `json:"signer" bson:"signer"`
	Signature string `json:"signature" bson:"signature"` // Assuming signature is a string representation
}
