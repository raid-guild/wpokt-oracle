package util

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const primaryType = "Message"

type DomainData struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}

var typesStandard = apitypes.Types{
	"EIP712Domain": {
		{
			Name: "name",
			Type: "string",
		},
		{
			Name: "version",
			Type: "string",
		},
		{
			Name: "chainId",
			Type: "uint256",
		},
		{
			Name: "verifyingContract",
			Type: "address",
		},
	},
	"Message": {
		{
			Name: "version",
			Type: "uint8",
		},
		{
			Name: "nonce",
			Type: "uint32",
		},
		{
			Name: "originDomain",
			Type: "uint32",
		},
		{
			Name: "sender",
			Type: "bytes32",
		},
		{
			Name: "destinationDomain",
			Type: "uint32",
		},
		{
			Name: "recipient",
			Type: "bytes32",
		},
		{
			Name: "messageBody",
			Type: "bytes",
		},
	},
}

func signTypedData(
	content models.MessageContent,
	domainData DomainData,
	key *ecdsa.PrivateKey,
) ([]byte, error) {

	message := apitypes.TypedDataMessage{
		"version":           content.Version,
		"nonce":             content.Nonce,
		"originDomain":      content.OriginDomain,
		"sender":            content.Sender,
		"destinationDomain": content.DestinationDomain,
		"recipient":         content.Recipient,
		"messageBody":       content.MessageBody,
	}

	domain := apitypes.TypedDataDomain{
		Name:              domainData.Name,
		Version:           domainData.Version,
		ChainId:           math.NewHexOrDecimal256(domainData.ChainId.Int64()),
		VerifyingContract: domainData.VerifyingContract.String(),
	}

	typedData := apitypes.TypedData{
		Types:       typesStandard,
		PrimaryType: primaryType,
		Domain:      domain,
		Message:     message,
	}

	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, err
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, err
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	sighash := crypto.Keccak256(rawData)

	signature, err := crypto.Sign(sighash, key)
	if signature[64] == 0 || signature[64] == 1 {
		signature[64] += 27
	}

	return signature, err
}

func SignMessage(
	message *models.Message,
	domain DomainData,
	privateKey *ecdsa.PrivateKey,
) error {
	signature, err := signTypedData(message.Content, domain, privateKey)
	if err != nil {
		return err
	}

	signatureEncoded := "0x" + hex.EncodeToString(signature)
	signatures := message.Signatures

	sig := models.Signature{
		Signer:    strings.ToLower(crypto.PubkeyToAddress(privateKey.PublicKey).Hex()),
		Signature: signatureEncoded,
	}
	signatures = append(signatures, sig)

	sort.Slice(signatures, func(i, j int) bool {
		return common.HexToAddress(signatures[i].Signer).Big().Cmp(common.HexToAddress(signatures[j].Signer).Big()) == -1
	})

	message.Signatures = signatures
	return nil
}
