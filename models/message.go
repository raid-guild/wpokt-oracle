package models

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/dan13ram/wpokt-oracle/common"
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

func (content *MessageContent) MessageID() []byte {
	return []byte{}
}

func (content *MessageContent) EncodeToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, content.Version); err != nil { // 1 byte
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, content.Nonce); err != nil { // 4 bytes
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, content.OriginDomain); err != nil { // 4 bytes
		return nil, err
	}
	senderBytes, err := common.HexAddressToBytes32(content.Sender)
	if err != nil {
		return nil, err
	}
	if _, err = buf.Write(senderBytes[:]); err != nil { // 32 bytes
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, content.DestinationDomain); err != nil { // 4 bytes
		return nil, err
	}
	recipientBytes, err := common.HexAddressToBytes32(content.Recipient)
	if err != nil {
		return nil, err
	}
	if _, err = buf.Write(recipientBytes[:]); err != nil { // 32 bytes
		return nil, err
	}
	bodyBytes, err := content.MessageBody.EncodeToBytes()
	if err != nil {
		return nil, err
	}
	if _, err = buf.Write(bodyBytes); err != nil { // 96 bytes
		return nil, err
	}

	// total 173 bytes

	return buf.Bytes(), nil
}

func (content *MessageContent) DecodeFromBytes(data []byte) error {
	*content = MessageContent{}

	if len(data) != 173 {
		return fmt.Errorf("invalid data length")
	}

	buf := bytes.NewReader(data)

	if err := binary.Read(buf, binary.BigEndian, &content.Version); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &content.Nonce); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &content.OriginDomain); err != nil {
		return err
	}
	senderBytes := make([]byte, 32)
	if _, err := io.ReadFull(buf, senderBytes); err != nil {
		return err
	}
	sender := "0x" + hex.EncodeToString(senderBytes[12:32])
	content.Sender = sender
	if err := binary.Read(buf, binary.BigEndian, &content.DestinationDomain); err != nil {
		return err
	}
	recipientBytes := make([]byte, 32)
	if _, err := io.ReadFull(buf, recipientBytes); err != nil {
		return err
	}
	recipient := "0x" + hex.EncodeToString(recipientBytes[12:32])
	content.Recipient = recipient
	bodyBytes := make([]byte, 96)
	if _, err := io.ReadFull(buf, bodyBytes); err != nil {
		return err
	}
	if err := content.MessageBody.DecodeFromBytes(bodyBytes); err != nil {
		return err
	}

	if buf.Len() != 0 {
		return fmt.Errorf("unexpected data")
	}

	return nil
}

type MessageStatus string

const (
	MessageStatusPending     MessageStatus = "pending"
	MessageStatusSigned      MessageStatus = "signed"
	MessageStatusBroadcasted MessageStatus = "broadcasted"
	MessageStatusSuccess     MessageStatus = "success"
	MessageStatusInvalid     MessageStatus = "invalid"
)

type Message struct {
	ID                    *primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	OriginTransaction     *primitive.ObjectID `json:"origin_transaction" bson:"origin_transaction"`
	OriginTransactionHash string              `json:"origin_transaction_hash" bson:"origin_transaction_hash"`
	MessageID             string              `json:"message_id" bson:"message_id"`
	Content               MessageContent      `json:"content" bson:"content"`
	Signatures            []Signature         `json:"signatures" bson:"signatures"`
	Transaction           primitive.ObjectID  `json:"transaction" bson:"transaction"`
	Sequence              uint64              `json:"sequence" bson:"sequence"` // account sequence for submitting the transaction
	Status                MessageStatus       `json:"status" bson:"status"`
	TransactionHash       string              `json:"transaction_hash" bson:"transaction_hash"`
	CreatedAt             time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time           `bson:"updated_at" json:"updated_at"`
}

type MessageBody struct {
	SenderAddress    string `json:"sender_address" bson:"sender_address"`
	Amount           uint64 `json:"amount" bson:"amount"`
	RecipientAddress string `json:"recipient_address" bson:"recipient_address"`
}

func (body *MessageBody) EncodeToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)

	recipientBytes, err := common.HexAddressToBytes32(body.RecipientAddress)
	if err != nil {
		return nil, err
	}
	if _, err = buf.Write(recipientBytes[:]); err != nil { // 32 bytes
		return nil, err
	}
	amount := new(big.Int).SetUint64(body.Amount)
	amountBytes := amount.FillBytes(make([]byte, 32))
	if _, err = buf.Write(amountBytes); err != nil { // 32 bytes
		return nil, err
	}
	senderBytes, err := common.HexAddressToBytes32(body.SenderAddress)
	if err != nil {
		return nil, err
	}
	if _, err := buf.Write(senderBytes[:]); err != nil { // 32 bytes
		return nil, err
	}

	// total 96 bytes

	return buf.Bytes(), nil
}

func (body *MessageBody) DecodeFromBytes(data []byte) error {
	*body = MessageBody{}
	if len(data) != 96 {
		return fmt.Errorf("invalid data length")
	}

	buf := bytes.NewReader(data)

	var recipientBytes [32]byte
	if _, err := io.ReadFull(buf, recipientBytes[:]); err != nil {
		return err
	}

	amountBytes := make([]byte, 32)
	if _, err := io.ReadFull(buf, amountBytes); err != nil {
		return err
	}
	amount := new(big.Int).SetBytes(amountBytes)

	var senderBytes [32]byte
	if _, err := io.ReadFull(buf, senderBytes[:]); err != nil {
		return err
	}

	if buf.Len() != 0 {
		return fmt.Errorf("unexpected data")
	}

	recipient := "0x" + hex.EncodeToString(recipientBytes[12:32])
	sender := "0x" + hex.EncodeToString(senderBytes[12:32])

	*body = MessageBody{
		RecipientAddress: recipient,
		Amount:           amount.Uint64(),
		SenderAddress:    sender,
	}

	return nil
}

type Signature struct {
	Signer    string `json:"signer" bson:"signer"`
	Signature string `json:"signature" bson:"signature"` // Assuming signature is a string representation
}
