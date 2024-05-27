package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefundStatus string

const (
	RefundStatusPending RefundStatus = "pending"
	RefundStatusSigned  RefundStatus = "signed"
	RefundStatusSuccess RefundStatus = "success"
	RefundStatusInvalid RefundStatus = "invalid"
)

type Refund struct {
	ID                    *primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	OriginTransaction     *primitive.ObjectID `json:"origin_transaction" bson:"origin_transaction"`
	OriginTransactionHash string              `json:"origin_transaction_hash" bson:"origin_transaction_hash"`
	Recipient             string              `json:"recipient" bson:"recipient"`
	Amount                uint64              `json:"amount" bson:"amount"`
	TransactionBody       string              `json:"transaction_body" bson:"transaction_body"`
	Signatures            []Signature         `json:"signatures" bson:"signatures"`
	Transaction           *primitive.ObjectID `json:"transaction" bson:"transaction"`
	TransactionHash       string              `json:"transaction_hash" bson:"transaction_hash"`
	Status                RefundStatus        `json:"status" bson:"status"`
	CreatedAt             time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time           `bson:"updated_at" json:"updated_at"`
}
