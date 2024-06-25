package cosmos

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/cosmos/cosmos-sdk/client"

	clientMocks "github.com/dan13ram/wpokt-oracle/cosmos/client/mocks"
	"github.com/dan13ram/wpokt-oracle/models"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	ethcommon "github.com/ethereum/go-ethereum/common"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/dan13ram/wpokt-oracle/common"
)

func TestCosmosSignTx(t *testing.T) {
	mockClient := clientMocks.NewMockCosmosClient(t)

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	config := models.CosmosNetworkConfig{
		ChainID:         "chain-id",
		CoinDenom:       "upokt",
		Bech32Prefix:    "pokt",
		MultisigAddress: multisigAddr,
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)
	tx := clientMocks.NewMockTx(t)

	utilNewSendTx = func(string, []byte, []byte, sdk.Coin, string, sdk.Coin) (string, error) {
		return "txBody", nil
	}

	utilWrapTxBuilder = func(prefix string, txBody string) (client.TxBuilder, client.TxConfig, error) {
		assert.Equal(t, "pokt", prefix)
		assert.Equal(t, "txBody", txBody)
		return txBuilder, txConfig, nil
	}

	utilSignWithPrivKey = func(context.Context, signing.SignerData, client.TxBuilder, cryptotypes.PrivKey, client.TxConfig, uint64) (signingtypes.SignatureV2, []byte, error) {
		return signingtypes.SignatureV2{
			PubKey: signerKey.PubKey(),
			Data: &signingtypes.SingleSignatureData{
				SignMode:  signingtypes.SignMode_SIGN_MODE_DIRECT,
				Signature: []byte("signature"),
			},
		}, nil, nil
	}

	signers := [][]byte{multisigPk.Address().Bytes()}

	assert.True(t, isTxSigner(multisigPk.Address().Bytes(), signers))

	tx.EXPECT().GetSigners().Return(signers, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)
	txBuilder.EXPECT().GetTx().Return(tx)

	var encoder sdk.TxEncoder = func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	}

	txConfig.EXPECT().TxJSONEncoder().Return(encoder)

	mockClient.EXPECT().GetAccount(multisigAddr).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	amount, _ := sdk.ParseCoinNormalized("100upokt")

	txBody, signatures, err := cosmosSignTx(
		signerKey,
		config,
		mockClient,
		*message.Sequence,
		message.Signatures,
		message.TransactionBody,
		recipientAddr[:],
		amount,
		"memo",
	)

	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)

	assert.Nil(t, err)
	assert.Equal(t, "encoded tx", txBody)
	assert.NotEmpty(t, signatures)
}
