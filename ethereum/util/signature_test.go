package util

import (
	"math/big"
	"strings"
	"testing"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestHexToBytes32(t *testing.T) {
	hexString := "0xAb5801a7D398351b8bE11C439e05C5b3259aec9B"
	expected := "0x000000000000000000000000ab5801a7d398351b8be11c439e05c5b3259aec9b"
	result := HexToBytes32(hexString)
	assert.Equal(t, expected, result)
}

func TestSignTypedData(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	content := models.MessageContent{
		Version:           1,
		Nonce:             1,
		OriginDomain:      1,
		Sender:            "0xAb5801a7D398351b8bE11C439e05C5b3259aec9B",
		DestinationDomain: 1,
		Recipient:         "0xAb5801a7D398351b8bE11C439e05C5b3259aec9B",
		MessageBody: models.MessageBody{
			SenderAddress:    "0xAb5801a7D398351b8bE11C439e05C5b3259aec9B",
			Amount:           "100",
			RecipientAddress: "0xAb5801a7D398351b8bE11C439e05C5b3259aec9B",
		},
	}

	domain := DomainData{
		Name:              "Test",
		Version:           "1",
		ChainId:           big.NewInt(1),
		VerifyingContract: common.HexToAddress("0xAb5801a7D398351b8bE11C439e05C5b3259aec9B"),
	}

	signature, err := signTypedData(content, domain, privateKey)
	assert.NoError(t, err)
	assert.NotNil(t, signature)
}

func TestSignMessage(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	message := &models.Message{
		Content: models.MessageContent{
			Version:           1,
			Nonce:             1,
			OriginDomain:      1,
			Sender:            "0xAb5801a7D398351b8bE11C439e05C5b3259aec9B",
			DestinationDomain: 1,
			Recipient:         "0xAb5801a7D398351b8bE11C439e05C5b3259aec9B",
			MessageBody: models.MessageBody{
				SenderAddress:    "0xAb5801a7D398351b8bE11C439e05C5b3259aec9B",
				Amount:           "100",
				RecipientAddress: "0xAb5801a7D398351b8bE11C439e05C5b3259aec9B",
			},
		},
	}

	domain := DomainData{
		Name:              "Test",
		Version:           "1",
		ChainId:           big.NewInt(1),
		VerifyingContract: common.HexToAddress("0xAb5801a7D398351b8bE11C439e05C5b3259aec9B"),
	}

	err = SignMessage(message, domain, privateKey)
	assert.NoError(t, err)
	assert.Len(t, message.Signatures, 1)
	assert.Equal(t, strings.ToLower(crypto.PubkeyToAddress(privateKey.PublicKey).Hex()), message.Signatures[0].Signer)
}
