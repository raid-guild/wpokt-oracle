package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TxStatus string

const (
	TxStatusPending   TxStatus = "pending"
	TxStatusConfirmed TxStatus = "confirmed"
	TxStatusSuccess   TxStatus = "success"
	TxStatusFailed    TxStatus = "failed"
	TxStatusInvalid   TxStatus = "invalid"
)

type Transaction struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Hash          []byte             `json:"hash" bson:"hash"`
	Sender        []byte             `json:"sender" bson:"sender"`
	BlockHeight   uint64             `json:"block_height" bson:"block_height"`
	Confirmations uint64             `json:"confirmations" bson:"confirmations"`
	Chain         Chain              `bson:"chain" json:"chain"`
	TxStatus      TxStatus           `json:"tx_status" bson:"tx_status"`
	Refund        *RefundInfo        `json:"refund" bson:"refund"`
	MessageID     primitive.ObjectID `json:"message_id" bson:"message_id"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

type RefundInfo struct {
	Required     bool   `json:"required" bson:"required"`
	Refunded     bool   `json:"refunded" bson:"refunded"`
	RefundAmount uint64 `json:"refund_amount" bson:"refund_amount"`
	RefundTxHash []byte `json:"refund_tx_hash" bson:"refund_tx_hash"`
}

type MintMemo struct {
	Address string `json:"address" bson:"address"`
	ChainID string `json:"chain_id" bson:"chain_id"`
}
