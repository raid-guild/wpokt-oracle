package util

import (
	"context"
	"testing"

	// "github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/crypto/hd"
	// "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	// cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/stretchr/testify/assert"
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

// func TestSignWithPrivKey_ErrorSigning(t *testing.T) {
// 	// Generate a new private key
// 	privKey := secp256k1.GenPrivKey()
//
// 	// Create a new TxConfig
// 	txConfig := NewTxConfig("pokt")
//
// 	// Create a new TxBuilder
// 	txBuilder := txConfig.NewTxBuilder()
//
// 	// Create dummy signer data
// 	signerData := authsigning.SignerData{
// 		ChainID:       "poktroll",
// 		AccountNumber: 1,
// 		Sequence:      1,
// 		PubKey:        privKey.PubKey(),
// 		Address:       sdk.AccAddress(privKey.PubKey().Address()).String(),
// 	}
//
// 	// Create a new context
// 	ctx := context.Background()
//
// 	// Call the SignWithPrivKey function with an invalid private key to trigger an error
// 	privKey = &secp256k1.PrivKey{Key: []byte{}}
// 	sigV2, _, err := SignWithPrivKey(ctx, signerData, txBuilder, privKey, txConfig, 1)
//
// 	// Assertions
// 	assert.Error(t, err)
// 	assert.Empty(t, sigV2)
// }
