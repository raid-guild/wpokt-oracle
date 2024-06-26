package util

import (
	"context"
	"errors"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	clientMocks "github.com/dan13ram/wpokt-oracle/cosmos/client/mocks"
	"github.com/dan13ram/wpokt-oracle/cosmos/util/mocks"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSignWithPrivKey(t *testing.T) {
	// Generate a new private key
	privKey := secp256k1.GenPrivKey()

	// Create a new TxConfig
	txConfig := NewTxConfig("pokt")

	// Create a new TxBuilder
	txBuilder := txConfig.NewTxBuilder()

	// Create dummy signer data
	signerData := authsigning.SignerData{
		ChainID:       "poktroll",
		AccountNumber: 1,
		Sequence:      1,
		PubKey:        privKey.PubKey(),
		Address:       sdk.AccAddress(privKey.PubKey().Address()).String(),
	}

	// Create a new context
	ctx := context.Background()

	// Call the SignWithPrivKey function
	sigV2, msg, err := SignWithPrivKey(ctx, signerData, txBuilder, privKey, txConfig, 1)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, sigV2)
	assert.Equal(t, privKey.PubKey(), sigV2.PubKey)
	assert.Equal(t, signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, sigV2.Data.(*signingtypes.SingleSignatureData).SignMode)
	assert.NotEmpty(t, sigV2.Data.(*signingtypes.SingleSignatureData).Signature)
	assert.Equal(t, uint64(1), sigV2.Sequence)
	pub := privKey.PubKey()
	assert.Equal(t, true, pub.VerifySignature(msg, sigV2.Data.(*signingtypes.SingleSignatureData).Signature))
}

func TestSignWithPrivKey_ErrorSignBytes(t *testing.T) {
	// Generate a new private key
	privKey := secp256k1.GenPrivKey()

	// Create a new TxConfig
	txConfig := NewTxConfig("pokt")

	// Create a new TxBuilder
	txBuilder := txConfig.NewTxBuilder()

	pubKey := &secp256k1.PubKey{}

	// Create dummy signer data
	signerData := authsigning.SignerData{
		ChainID:       "poktroll",
		AccountNumber: 1,
		Sequence:      1,
		PubKey:        pubKey,
	}

	// Create a new context
	ctx := context.Background()

	// Call the SignWithPrivKey function
	sigV2, _, err := SignWithPrivKey(ctx, signerData, txBuilder, privKey, txConfig, 1)

	// Assertions
	assert.Error(t, err)
	assert.Empty(t, sigV2)
}

func TestSignWithPrivKey_ErrorSigning(t *testing.T) {
	// Generate a new private key
	privKey := secp256k1.GenPrivKey()

	// Create a new TxConfig
	txConfig := NewTxConfig("pokt")

	// Create a new TxBuilder
	txBuilder := txConfig.NewTxBuilder()

	// Create dummy signer data
	signerData := authsigning.SignerData{
		ChainID:       "poktroll",
		AccountNumber: 1,
		Sequence:      1,
		PubKey:        privKey.PubKey(),
		Address:       sdk.AccAddress(privKey.PubKey().Address()).String(),
	}

	// Create a new context
	ctx := context.Background()

	mockPrivKey := mocks.NewMockPrivKey(t)
	mockPrivKey.EXPECT().Sign(mock.Anything).Return(nil, errors.New("error signing"))

	// Call the SignWithPrivKey function
	sigV2, _, err := SignWithPrivKey(ctx, signerData, txBuilder, mockPrivKey, txConfig, 1)

	// Assertions
	assert.Error(t, err)
	assert.Empty(t, sigV2)
}

func TestValidateSignature(t *testing.T) {
	privKey := secp256k1.GenPrivKey()

	// Create a new TxConfig
	txConfig := NewTxConfig("pokt")

	// Create a new TxBuilder
	txBuilder := txConfig.NewTxBuilder()

	// Create dummy signer data
	signerData := authsigning.SignerData{
		ChainID:       "poktroll",
		AccountNumber: 1,
		Sequence:      1,
		PubKey:        privKey.PubKey(),
		Address:       sdk.AccAddress(privKey.PubKey().Address()).String(),
	}

	// Create a new context
	ctx := context.Background()

	// Call the SignWithPrivKey function
	sigV2, _, err := SignWithPrivKey(ctx, signerData, txBuilder, privKey, txConfig, 1)
	assert.NoError(t, err)

	config := models.CosmosNetworkConfig{
		ChainID:      "poktroll",
		Bech32Prefix: "pokt",
	}

	err = ValidateSignature(config, &sigV2, 1, 1, txConfig, txBuilder)
	assert.NoError(t, err)
}

func TestValidateSignature_VerificationFailure(t *testing.T) {
	privKey := secp256k1.GenPrivKey()

	// Create a new TxConfig
	txConfig := NewTxConfig("pokt")

	// Create a new TxBuilder
	txBuilder := txConfig.NewTxBuilder()

	// Create dummy signer data
	signerData := authsigning.SignerData{
		ChainID:       "poktroll",
		AccountNumber: 1,
		Sequence:      1,
		PubKey:        privKey.PubKey(),
		Address:       sdk.AccAddress(privKey.PubKey().Address()).String(),
	}

	// Create a new context
	ctx := context.Background()

	// Call the SignWithPrivKey function
	sigV2, _, err := SignWithPrivKey(ctx, signerData, txBuilder, privKey, txConfig, 1)
	assert.NoError(t, err)

	config := models.CosmosNetworkConfig{
		ChainID:      "poktroll-different",
		Bech32Prefix: "pokt",
	}

	err = ValidateSignature(config, &sigV2, 1, 1, txConfig, txBuilder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "couldn't verify signature for address")
}

func TestValidateSignature_AnyError(t *testing.T) {
	privKey := secp256k1.GenPrivKey()

	// Create a new TxConfig
	txConfig := NewTxConfig("pokt")

	// Create a new TxBuilder
	txBuilder := txConfig.NewTxBuilder()

	// Create dummy signer data
	signerData := authsigning.SignerData{
		ChainID:       "poktroll",
		AccountNumber: 1,
		Sequence:      1,
		PubKey:        privKey.PubKey(),
		Address:       sdk.AccAddress(privKey.PubKey().Address()).String(),
	}

	// Create a new context
	ctx := context.Background()

	// Call the SignWithPrivKey function
	sigV2, _, err := SignWithPrivKey(ctx, signerData, txBuilder, privKey, txConfig, 1)
	assert.NoError(t, err)

	config := models.CosmosNetworkConfig{
		ChainID:      "poktroll-different",
		Bech32Prefix: "pokt",
	}

	sigV2.PubKey = nil

	err = ValidateSignature(config, &sigV2, 1, 1, txConfig, txBuilder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating any pubkey")
}

func TestValidateSignature_TxError(t *testing.T) {
	privKey := secp256k1.GenPrivKey()

	// Create a new TxConfig
	txConfig := NewTxConfig("pokt")

	// Create a new TxBuilder
	txBuilder := clientMocks.NewMockTxBuilder(t)

	tx := clientMocks.NewMockTx(t)

	txBuilder.EXPECT().GetTx().Return(tx)

	// Call the SignWithPrivKey function
	sigV2 := signingtypes.SignatureV2{
		PubKey: privKey.PubKey(),
	}

	config := models.CosmosNetworkConfig{
		ChainID:      "poktroll",
		Bech32Prefix: "pokt",
	}

	err := ValidateSignature(config, &sigV2, 1, 1, txConfig, txBuilder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected Tx to be signing.V2AdaptableTx")
}
