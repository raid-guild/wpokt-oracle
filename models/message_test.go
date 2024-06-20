package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageContent_MessageID(t *testing.T) {
	content := MessageContent{
		Version:           1,
		Nonce:             12345,
		OriginDomain:      1,
		Sender:            "0x0000000000000000000000000000000000000001",
		DestinationDomain: 2,
		Recipient:         "0x0000000000000000000000000000000000000002",
		MessageBody: MessageBody{
			SenderAddress:    "0x0000000000000000000000000000000000000001",
			Amount:           "1000",
			RecipientAddress: "0x0000000000000000000000000000000000000002",
		},
	}

	messageID, err := content.MessageID()
	assert.NoError(t, err)
	assert.NotNil(t, messageID)
	assert.Equal(t, 32, len(messageID))
}

func TestMessageContent_EncodeToBytes(t *testing.T) {
	content := MessageContent{
		Version:           1,
		Nonce:             12345,
		OriginDomain:      1,
		Sender:            "0x0000000000000000000000000000000000000001",
		DestinationDomain: 2,
		Recipient:         "0x0000000000000000000000000000000000000002",
		MessageBody: MessageBody{
			SenderAddress:    "0x0000000000000000000000000000000000000001",
			Amount:           "1000",
			RecipientAddress: "0x0000000000000000000000000000000000000002",
		},
	}

	encoded, err := content.EncodeToBytes()
	assert.NoError(t, err)
	assert.NotNil(t, encoded)
	assert.Equal(t, 173, len(encoded))
}

func TestMessageContent_DecodeFromBytes(t *testing.T) {
	content := MessageContent{
		Version:           1,
		Nonce:             12345,
		OriginDomain:      1,
		Sender:            "0x0000000000000000000000000000000000000001",
		DestinationDomain: 2,
		Recipient:         "0x0000000000000000000000000000000000000002",
		MessageBody: MessageBody{
			SenderAddress:    "0x0000000000000000000000000000000000000001",
			Amount:           "1000",
			RecipientAddress: "0x0000000000000000000000000000000000000002",
		},
	}

	encoded, _ := content.EncodeToBytes()

	var decodedContent MessageContent
	err := decodedContent.DecodeFromBytes(encoded)
	assert.NoError(t, err)
	assert.Equal(t, content, decodedContent)
}

func TestMessageBody_EncodeToBytes(t *testing.T) {
	body := MessageBody{
		SenderAddress:    "0x0000000000000000000000000000000000000001",
		Amount:           "1000",
		RecipientAddress: "0x0000000000000000000000000000000000000002",
	}

	encoded, err := body.EncodeToBytes()
	assert.NoError(t, err)
	assert.NotNil(t, encoded)
	assert.Equal(t, 96, len(encoded))
}

func TestMessageBody_DecodeFromBytes(t *testing.T) {
	body := MessageBody{
		SenderAddress:    "0x0000000000000000000000000000000000000001",
		Amount:           "1000",
		RecipientAddress: "0x0000000000000000000000000000000000000002",
	}

	encoded, _ := body.EncodeToBytes()

	var decodedBody MessageBody
	err := decodedBody.DecodeFromBytes(encoded)
	assert.NoError(t, err)
	assert.Equal(t, body, decodedBody)
}

func TestMessageContent_DecodeFromBytes_InvalidLength(t *testing.T) {
	data := make([]byte, 100) // Invalid length
	var content MessageContent
	err := content.DecodeFromBytes(data)
	assert.Error(t, err)
	assert.Equal(t, "invalid data length", err.Error())
}

func TestMessageBody_DecodeFromBytes_InvalidLength(t *testing.T) {
	data := make([]byte, 50) // Invalid length
	var body MessageBody
	err := body.DecodeFromBytes(data)
	assert.Error(t, err)
	assert.Equal(t, "invalid data length", err.Error())
}

func TestMessage_DecodeFromBytes_UnexpectedData(t *testing.T) {
	content := MessageContent{
		Version:           1,
		Nonce:             12345,
		OriginDomain:      1,
		Sender:            "0x0000000000000000000000000000000000000001",
		DestinationDomain: 2,
		Recipient:         "0x0000000000000000000000000000000000000002",
		MessageBody: MessageBody{
			SenderAddress:    "0x0000000000000000000000000000000000000001",
			Amount:           "1000",
			RecipientAddress: "0x0000000000000000000000000000000000000002",
		},
	}

	encoded, _ := content.EncodeToBytes()
	encoded = append(encoded, 0x00) // Add unexpected data

	var decodedContent MessageContent
	err := decodedContent.DecodeFromBytes(encoded)
	assert.Error(t, err)
	assert.Equal(t, "invalid data length", err.Error())
}
