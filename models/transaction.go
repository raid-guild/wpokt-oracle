package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TxStatus string

const (
	TxStatusPending   TxStatus = "pending"
	TxStatusConfirmed TxStatus = "confirmed"
	TxStatusFailed    TxStatus = "failed"
	TxStatusInvalid   TxStatus = "invalid"
)

type Transaction struct {
	ID            *primitive.ObjectID `json:"id" bson:"_id"`
	Hash          string              `json:"hash" bson:"hash"`
	Sender        string              `json:"sender" bson:"sender"`
	BlockHeight   uint64              `json:"block_height" bson:"block_height"`
	Confirmations uint64              `json:"confirmations" bson:"confirmations"`
	Chain         Chain               `bson:"chain" json:"chain"`
	TxStatus      TxStatus            `json:"tx_status" bson:"tx_status"`
	CreatedAt     time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time           `bson:"updated_at" json:"updated_at"`
	Refund        *primitive.ObjectID `json:"refund" bson:"refund"`
	Message       *primitive.ObjectID `json:"message" bson:"message"`
}

type MintMemo struct {
	Address string `json:"address" bson:"address"`
	ChainID string `json:"chain_id" bson:"chain_id"`
}
