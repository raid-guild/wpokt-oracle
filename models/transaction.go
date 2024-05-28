package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusConfirmed TransactionStatus = "confirmed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusInvalid   TransactionStatus = "invalid"
)

type Transaction struct {
	ID            *primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Hash          string              `json:"hash" bson:"hash"`
	FromAddress   string              `json:"from_address" bson:"from_address"`
	ToAddress     string              `json:"to_address" bson:"to_address"`
	BlockHeight   uint64              `json:"block_height" bson:"block_height"`
	Confirmations uint64              `json:"confirmations" bson:"confirmations"`
	Chain         Chain               `bson:"chain" json:"chain"`
	Status        TransactionStatus   `json:"status" bson:"status"`
	CreatedAt     time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time           `bson:"updated_at" json:"updated_at"`
	Refund        *primitive.ObjectID `json:"refund" bson:"refund"`
	Message       *primitive.ObjectID `json:"message" bson:"message"`
}

type MintMemo struct {
	Address string `json:"address" bson:"address"`
	ChainID string `json:"chain_id" bson:"chain_id"`
}
