package util

import (
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"sort"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

const primaryType = "Message"

type DomainData struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract ethcommon.Address
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

func HexToBytes32(hexString string) string {
	bytes, _ := common.Bytes32FromAddressHex(hexString)
	return "0x" + hex.EncodeToString(bytes)
}

func signTypedData(
	content models.MessageContent,
	domainData DomainData,
	key *ecdsa.PrivateKey,
) ([]byte, error) {

	messageBodyBytes, err := content.MessageBody.EncodeToBytes()
	if err != nil {
		return nil, err
	}

	message := apitypes.TypedDataMessage{
		"version":           big.NewInt(int64(content.Version)),
		"nonce":             big.NewInt(int64(content.Nonce)),
		"originDomain":      big.NewInt(int64(content.OriginDomain)),
		"sender":            HexToBytes32(content.Sender),
		"destinationDomain": big.NewInt(int64(content.DestinationDomain)),
		"recipient":         HexToBytes32(content.Recipient),
		"messageBody":       messageBodyBytes,
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

	sighash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}

	signature, err := crypto.Sign(sighash, key)
	if err != nil {
		return nil, err
	}

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
		return ethcommon.HexToAddress(signatures[i].Signer).Big().Cmp(ethcommon.HexToAddress(signatures[j].Signer).Big()) == -1
	})

	message.Signatures = signatures
	return nil
}
