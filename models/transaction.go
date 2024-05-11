package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id"`
	Hash               []byte             `json:"hash" bson:"hash"`
	Timestamp          uint64             `json:"timestamp" bson:"timestamp"`
	Sender             []byte             `json:"sender" bson:"sender"`
	BlockHeight        uint64             `json:"blockHeight" bson:"blockHeight"`
	BlockConfirmations uint64             `json:"blockConfirmations" bson:"blockConfirmations"`
	ChainType          string             `json:"chainType" bson:"chainType"` // enum: ["cosmos", "ethereum"]
	ChainID            string             `json:"chainId" bson:"chainId"`
	ChainDomain        uint64             `json:"chainDomain" bson:"chainDomain"`
	TxStatus           string             `json:"txStatus" bson:"txStatus"` // enum: ["pending", "confirmed", "success", "failed"]
	IsValid            bool               `json:"isValid" bson:"isValid"`
	Refund             RefundInfo         `json:"refund" bson:"refund"`
	MessageID          primitive.ObjectID `json:"message_id" bson:"message_id"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
	UpdatedBy          primitive.ObjectID `bson:"updated_by" json:"updated_by"`
}

type RefundInfo struct {
	Required     bool   `json:"required" bson:"required"`
	Refunded     bool   `json:"refunded" bson:"refunded"`
	RefundTxHash []byte `json:"refundTxHash,omitempty" bson:"refundTxHash,omitempty"` // Optional: only if refund has been processed
}
