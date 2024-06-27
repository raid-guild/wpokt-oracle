package cosmos

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	multisigtypes "github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/anypb"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"

	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	clientMocks "github.com/dan13ram/wpokt-oracle/cosmos/client/mocks"
	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/db"
	dbMocks "github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/models"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	ethMocks "github.com/dan13ram/wpokt-oracle/ethereum/client/mocks"

	log "github.com/sirupsen/logrus"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	txsigning "cosmossdk.io/x/tx/signing"
)

func TestSignerHeight(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
	}

	height := signer.Height()

	assert.Equal(t, uint64(100), height)
}

func TestSignerUpdateCurrentHeight(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	signer.UpdateCurrentHeight()

	mockClient.AssertExpectations(t)
	assert.Equal(t, uint64(100), signer.currentBlockHeight)
}

func TestSignerUpdateCurrentHeight_Error(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), assert.AnError)

	signer.UpdateCurrentHeight()

	mockClient.AssertExpectations(t)
	assert.Equal(t, uint64(0), signer.currentBlockHeight)
}

func TestSignerUpdateMessage(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

	message := &models.Message{ID: &primitive.ObjectID{}}
	update := bson.M{"status": models.MessageStatusSigned}

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateMessage(message.ID, update).Return(nil)

	result := signer.UpdateMessage(message, update)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignerUpdateMessage_Error(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

	message := &models.Message{ID: &primitive.ObjectID{}}
	update := bson.M{"status": models.MessageStatusSigned}

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateMessage(message.ID, update).Return(assert.AnError)

	result := signer.UpdateMessage(message, update)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSign(t *testing.T) {
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	signer := &CosmosMessageSignerRunnable{
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:           "chain-id",
			CoinDenom:         "upokt",
			Bech32Prefix:      "pokt",
			MultisigAddress:   multisigAddr,
			MultisigThreshold: 4,
		},
	}

	coinAmount, _ := math.NewIntFromString("100")

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, nil
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	sequence := uint64(1)

	update, err := signer.Sign(
		&sequence,
		[]models.Signature{},
		"",
		recipientAddr[:],
		sdk.NewCoin("upokt", coinAmount),
		"Message",
	)

	expectUpdate := bson.M{
		"status":           models.MessageStatusPending,
		"transaction_body": "encoded tx",
		"signatures":       []models.Signature{{}},
		"sequence":         &sequence,
	}

	assert.NoError(t, err)
	assert.NotNil(t, update)
	assert.Equal(t, expectUpdate, update)
}

func TestSign_WithoutSequence(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)

	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:           "chain-id",
			CoinDenom:         "upokt",
			Bech32Prefix:      "pokt",
			MultisigAddress:   multisigAddr,
			MultisigThreshold: 4,
		},
	}

	coinAmount, _ := math.NewIntFromString("100")

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, nil
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	sequence := uint64(2)

	mockDB.EXPECT().LockReadSequences().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().FindMaxSequence(mock.Anything).Return(nil, nil)
	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: sequence}, nil)

	update, err := signer.Sign(
		nil,
		[]models.Signature{},
		"",
		recipientAddr[:],
		sdk.NewCoin("upokt", coinAmount),
		"Message",
	)

	expectUpdate := bson.M{
		"status":           models.MessageStatusPending,
		"transaction_body": "encoded tx",
		"signatures":       []models.Signature{{}},
		"sequence":         &sequence,
	}

	assert.NoError(t, err)
	assert.NotNil(t, update)
	assert.Equal(t, expectUpdate, update)
}

func TestSign_WithoutSequence_Error(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)

	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:           "chain-id",
			CoinDenom:         "upokt",
			Bech32Prefix:      "pokt",
			MultisigAddress:   multisigAddr,
			MultisigThreshold: 4,
		},
	}

	coinAmount, _ := math.NewIntFromString("100")

	mockDB.EXPECT().LockReadSequences().Return("lock-id", assert.AnError)

	update, err := signer.Sign(
		nil,
		[]models.Signature{},
		"",
		recipientAddr[:],
		sdk.NewCoin("upokt", coinAmount),
		"Message",
	)

	assert.Error(t, err)
	assert.Nil(t, update)
}

func TestSign_AboveThreshold(t *testing.T) {
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	signer := &CosmosMessageSignerRunnable{
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:           "chain-id",
			CoinDenom:         "upokt",
			Bech32Prefix:      "pokt",
			MultisigAddress:   multisigAddr,
			MultisigThreshold: 1,
		},
	}

	coinAmount, _ := math.NewIntFromString("100")

	signs := []models.Signature{{}, {}}

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx with signatures", signs, nil
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	sequence := uint64(1)

	assert.True(t, len(signs) >= int(signer.config.MultisigThreshold))

	update, err := signer.Sign(
		&sequence,
		[]models.Signature{},
		"",
		recipientAddr[:],
		sdk.NewCoin("upokt", coinAmount),
		"Message",
	)

	expectUpdate := bson.M{
		"status":           models.MessageStatusSigned,
		"transaction_body": "encoded tx with signatures",
		"signatures":       signs,
		"sequence":         &sequence,
	}

	assert.NoError(t, err)
	assert.NotNil(t, update)
	assert.Equal(t, expectUpdate, update)
}

func TestSign_AlreadySigned(t *testing.T) {
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	signer := &CosmosMessageSignerRunnable{
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:           "chain-id",
			CoinDenom:         "upokt",
			Bech32Prefix:      "pokt",
			MultisigAddress:   multisigAddr,
			MultisigThreshold: 4,
		},
	}

	coinAmount, _ := math.NewIntFromString("100")

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, ErrAlreadySigned
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	sequence := uint64(1)

	update, err := signer.Sign(
		&sequence,
		[]models.Signature{},
		"",
		recipientAddr[:],
		sdk.NewCoin("upokt", coinAmount),
		"Message",
	)

	assert.Error(t, err)
	assert.Nil(t, update)
	assert.Equal(t, ErrAlreadySigned, err)
}

func TestSignMessage_AddressError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: "recipientAddr", Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	result := signer.SignMessage(message)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignMessage_AmountError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "invalid"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	result := signer.SignMessage(message)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignMessage_AlreadySigned(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, ErrAlreadySigned
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	result := signer.SignMessage(message)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignMessage_ErrorSigning(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, assert.AnError
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	result := signer.SignMessage(message)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignMessage_ErrorLocking(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", assert.AnError)

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, nil
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	result := signer.SignMessage(message)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignMessage_ErrorUpdating(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, nil
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(assert.AnError)

	result := signer.SignMessage(message)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignMessage(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, nil
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	result := signer.SignMessage(message)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignRefund_InvalidRecipient(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             "recipient",
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	result := signer.SignRefund(refund)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignRefund_InvalidAmount(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "invalid",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	result := signer.SignRefund(refund)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignRefund_AlreadySigned(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, ErrAlreadySigned
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	result := signer.SignRefund(refund)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignRefund_ErrorSigning(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, assert.AnError
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	result := signer.SignRefund(refund)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignRefund_LockError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", assert.AnError)

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, nil
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	result := signer.SignRefund(refund)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignRefund_UpdateError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, nil
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	mockDB.EXPECT().UpdateRefund(refund.ID, mock.Anything).Return(assert.AnError)

	result := signer.SignRefund(refund)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignRefund(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, nil
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	mockDB.EXPECT().UpdateRefund(refund.ID, mock.Anything).Return(nil)

	result := signer.SignRefund(refund)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestValidateAndFindDispatchIDEvent_InvalidMessageID(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             "invalid",
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	assert.Nil(t, result)
	assert.Error(t, err)
	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
}

func TestValidateAndFindDispatchIDEvent_EthClientError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestValidateAndFindDispatchIDEvent_MailboxError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestValidateAndFindDispatchIDEvent_ReceiptError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(nil, assert.AnError)

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestValidateAndFindDispatchIDEvent_UnsuccessfulReceipt(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusFailed,
		Logs:        []*types.Log{},
	}, nil).Once()

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	assert.NotNil(t, result)
	assert.Equal(t, models.TransactionStatusFailed, result.TxStatus)
	assert.NoError(t, err)

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(nil, nil).Once()

	result, err = signer.ValidateAndFindDispatchIDEvent(message)

	assert.NotNil(t, result)
	assert.Equal(t, models.TransactionStatusFailed, result.TxStatus)
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
}

func TestValidateAndFindDispatchIDEvent_BlockHeightError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), assert.AnError)
	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestValidateAndFindDispatchIDEvent_NoEvent(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(nil, assert.AnError)

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), result.Confirmations)
	assert.Equal(t, models.TransactionStatusInvalid, result.TxStatus)
}

func TestValidateAndFindDispatchIDEvent(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Equal(t, messageIDBytes, result.Event.MessageId)
	assert.Equal(t, uint64(10), result.Confirmations)
	assert.Equal(t, models.TransactionStatusConfirmed, result.TxStatus)
}

func TestValidateEthereumTxAndSignMessage_ValidateError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             "0xinvalid",
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	result := signer.ValidateEthereumTxAndSignMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateEthereumTxAndSignMessage_Pending(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(91), nil)
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	result := signer.ValidateEthereumTxAndSignMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateEthereumTxAndSignMessage_FailedTx(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusFailed,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)

	mockDB.EXPECT().UpdateMessage(message.ID, bson.M{"status": models.MessageStatusInvalid}).Return(nil)

	result := signer.ValidateEthereumTxAndSignMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.True(t, result)
}

func TestValidateEthereumTxAndSignMessage_LockError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	mockDB.EXPECT().LockWriteMessage(mock.Anything).Return("lock-id", assert.AnError)

	result := signer.ValidateEthereumTxAndSignMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateEthereumTxAndSignMessage(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)
	tx := clientMocks.NewMockTx(t)

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	utilNewSendTx = func(string, []byte, []byte, sdk.Coin, string, sdk.Coin) (string, error) {
		return "txBody", nil
	}

	utilWrapTxBuilder = func(prefix string, txBody string) (client.TxBuilder, client.TxConfig, error) {
		assert.Equal(t, "pokt", prefix)
		assert.Equal(t, "txBody", txBody)
		return txBuilder, txConfig, nil
	}

	utilSignWithPrivKey = func(context.Context, authsigning.SignerData, client.TxBuilder, crypto.PrivKey, client.TxConfig, uint64) (signingtypes.SignatureV2, []byte, error) {
		return signingtypes.SignatureV2{
			PubKey: signerKey.PubKey(),
			Data: &signingtypes.SingleSignatureData{
				SignMode:  signingtypes.SignMode_SIGN_MODE_DIRECT,
				Signature: []byte("signature"),
			},
		}, nil, nil
	}

	defer func() {
		utilNewSendTx = util.NewSendTx
		utilWrapTxBuilder = util.WrapTxBuilder
		utilSignWithPrivKey = util.SignWithPrivKey
	}()

	signers := [][]byte{multisigPk.Address().Bytes()}

	assert.True(t, isTxSigner(multisigPk.Address().Bytes(), signers))

	tx.EXPECT().GetSigners().Return(signers, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)
	txBuilder.EXPECT().GetTx().Return(tx)

	var encoder sdk.TxEncoder = func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	}

	txConfig.EXPECT().TxJSONEncoder().Return(encoder)

	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	mockClient.EXPECT().GetAccount(multisigAddr).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	mockDB.EXPECT().LockWriteMessage(mock.Anything).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	result := signer.ValidateEthereumTxAndSignMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignMessages(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)
	tx := clientMocks.NewMockTx(t)

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	utilNewSendTx = func(string, []byte, []byte, sdk.Coin, string, sdk.Coin) (string, error) {
		return "txBody", nil
	}

	utilWrapTxBuilder = func(prefix string, txBody string) (client.TxBuilder, client.TxConfig, error) {
		assert.Equal(t, "pokt", prefix)
		assert.Equal(t, "txBody", txBody)
		return txBuilder, txConfig, nil
	}

	utilSignWithPrivKey = func(context.Context, authsigning.SignerData, client.TxBuilder, crypto.PrivKey, client.TxConfig, uint64) (signingtypes.SignatureV2, []byte, error) {
		return signingtypes.SignatureV2{
			PubKey: signerKey.PubKey(),
			Data: &signingtypes.SingleSignatureData{
				SignMode:  signingtypes.SignMode_SIGN_MODE_DIRECT,
				Signature: []byte("signature"),
			},
		}, nil, nil
	}

	defer func() {
		utilNewSendTx = util.NewSendTx
		utilWrapTxBuilder = util.WrapTxBuilder
		utilSignWithPrivKey = util.SignWithPrivKey
	}()

	signers := [][]byte{multisigPk.Address().Bytes()}

	assert.True(t, isTxSigner(multisigPk.Address().Bytes(), signers))

	tx.EXPECT().GetSigners().Return(signers, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)
	txBuilder.EXPECT().GetTx().Return(tx)

	var encoder sdk.TxEncoder = func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	}

	txConfig.EXPECT().TxJSONEncoder().Return(encoder)

	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	mockClient.EXPECT().GetAccount(multisigAddr).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	mockDB.EXPECT().LockWriteMessage(mock.Anything).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	mockDB.EXPECT().GetPendingMessages(mock.Anything, mock.Anything).Return([]models.Message{*message}, nil)

	result := signer.SignMessages()

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignMessages_DBError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	signer := &CosmosMessageSignerRunnable{
		db:        mockDB,
		client:    mockClient,
		logger:    logger,
		signerKey: signerKey,
	}

	mockDB.EXPECT().GetPendingMessages(mock.Anything, mock.Anything).Return(nil, assert.AnError)

	result := signer.SignMessages()

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignerUpdateRefund(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

	refund := &models.Refund{ID: &primitive.ObjectID{}}
	update := bson.M{"status": models.RefundStatusSigned}

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateRefund(refund.ID, update).Return(nil)

	result := signer.UpdateRefund(refund, update)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignerUpdateRefund_Error(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

	refund := &models.Refund{ID: &primitive.ObjectID{}}
	update := bson.M{"status": models.RefundStatusSigned}

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateRefund(refund.ID, update).Return(assert.AnError)

	result := signer.UpdateRefund(refund, update)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestIsTxSigner(t *testing.T) {
	signer := secp256k1.GenPrivKey()
	signerPk := signer.PubKey()

	signers := [][]byte{signerPk.Bytes()}

	assert.True(t, isTxSigner(signerPk.Bytes(), signers))

	otherSigner := secp256k1.GenPrivKey()

	assert.False(t, isTxSigner(otherSigner.PubKey().Bytes(), signers))
}

func TestValidateSignatures_GetSignaturesV2Error(t *testing.T) {
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})

	signer := &CosmosMessageSignerRunnable{
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	tx := clientMocks.NewMockTx(t)

	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return(nil, assert.AnError)

	result := signer.ValidateSignaturesAndAddMultiSignatureToTxConfig("hash1", 1, txConfig, txBuilder)

	mockClient.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateSignatures_ThresholdError(t *testing.T) {
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})

	signer := &CosmosMessageSignerRunnable{
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			MultisigThreshold: 2,
		},
	}

	tx := clientMocks.NewMockTx(t)

	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	result := signer.ValidateSignaturesAndAddMultiSignatureToTxConfig("hash1", 1, txConfig, txBuilder)

	mockClient.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateSignatures_AccountError(t *testing.T) {
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})

	signer := &CosmosMessageSignerRunnable{
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	mockClient.EXPECT().GetAccount(mock.Anything).Return(nil, assert.AnError)

	tx := clientMocks.NewMockTx(t)

	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)

	result := signer.ValidateSignaturesAndAddMultiSignatureToTxConfig("hash1", 1, txConfig, txBuilder)

	mockClient.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateSignatures_Threshold(t *testing.T) {
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	signer := &CosmosMessageSignerRunnable{
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		chain: models.Chain{
			ChainID: "poktroll",
		},
		config: models.CosmosNetworkConfig{
			ChainID:           "poktroll",
			Bech32Prefix:      "pokt",
			MultisigThreshold: 1,
			MultisigAddress:   multisigAddr,
		},
	}

	mockClient.EXPECT().GetAccount(multisigAddr).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	// Create a new TxConfig
	txConfig := util.NewTxConfig("pokt")

	// Create a new TxBuilder
	txBuilder := txConfig.NewTxBuilder()

	// Create dummy signer data
	signerData := authsigning.SignerData{
		ChainID:       "poktroll",
		AccountNumber: 1,
		Sequence:      1,
		PubKey:        signerKey.PubKey(),
		Address:       sdk.AccAddress(signerKey.PubKey().Address()).String(),
	}

	sigV2, msg, err := util.SignWithPrivKey(context.Background(), signerData, txBuilder, signerKey, txConfig, 1)
	assert.NoError(t, err)

	err = txBuilder.SetSignatures(sigV2)
	assert.NoError(t, err)
	sigs, err := txBuilder.GetTx().GetSignaturesV2()
	assert.NoError(t, err)

	assert.Equal(t, sigV2, sigs[0])

	assert.True(t, signerKey.PubKey().VerifySignature(msg, sigV2.Data.(*signingtypes.SingleSignatureData).Signature))

	anyPk, err := codectypes.NewAnyWithValue(sigV2.PubKey)
	if err != nil {
		t.Errorf("error creating Any PB: %s", err)
	}
	txSignerData := txsigning.SignerData{
		ChainID:       "poktroll",
		AccountNumber: 1,
		Sequence:      1,
		Address:       sdk.AccAddress(sigV2.PubKey.Address()).String(),
		PubKey: &anypb.Any{
			TypeUrl: anyPk.TypeUrl,
			Value:   anyPk.Value,
		},
	}

	builtTx := txBuilder.GetTx()
	adaptableTx, ok := builtTx.(authsigning.V2AdaptableTx)
	if !ok {
		t.Errorf("expected Tx to be authsigning.V2AdaptableTx, got %T", builtTx)
	}

	txData := adaptableTx.GetSigningTxData()
	err = authsigning.VerifySignature(context.Background(), sigV2.PubKey, txSignerData, sigV2.Data,
		txConfig.SignModeHandler(), txData)

	assert.NoError(t, err)

	multisigSig := multisigtypes.NewMultisig(len(multisigPk.PubKeys))

	err = multisigtypes.AddSignatureV2(multisigSig, sigV2, multisigPk.GetPubKeys())
	assert.NoError(t, err)

	expectedSignature := signingtypes.SignatureV2{
		PubKey:   multisigPk,
		Data:     multisigSig,
		Sequence: 1,
	}

	result := signer.ValidateSignaturesAndAddMultiSignatureToTxConfig("hash1", 1, txConfig, txBuilder)
	assert.True(t, result)

	actualSigs, err := txBuilder.GetTx().GetSignaturesV2()
	assert.NoError(t, err)

	assert.Equal(t, expectedSignature, actualSigs[0])

	mockClient.AssertExpectations(t)
}

func TestValidateSignatures_Threshold_AnyError(t *testing.T) {
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	signer := &CosmosMessageSignerRunnable{
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		chain: models.Chain{
			ChainID: "poktroll",
		},
		config: models.CosmosNetworkConfig{
			ChainID:           "poktroll",
			Bech32Prefix:      "pokt",
			MultisigThreshold: 1,
			MultisigAddress:   multisigAddr,
		},
	}

	mockClient.EXPECT().GetAccount(multisigAddr).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txConfig := util.NewTxConfig("pokt")
	txBuilder := txConfig.NewTxBuilder()

	signerData := authsigning.SignerData{
		ChainID:       "poktroll",
		AccountNumber: 1,
		Sequence:      1,
		PubKey:        signerKey.PubKey(),
		Address:       sdk.AccAddress(signerKey.PubKey().Address()).String(),
	}

	sigV2, _, err := util.SignWithPrivKey(context.Background(), signerData, txBuilder, signerKey, txConfig, 1)
	assert.NoError(t, err)

	sigV2.PubKey = nil

	err = txBuilder.SetSignatures(sigV2)
	assert.NoError(t, err)

	result := signer.ValidateSignaturesAndAddMultiSignatureToTxConfig("hash1", 1, txConfig, txBuilder)
	assert.False(t, result)

	mockClient.AssertExpectations(t)
}

func TestValidateSignatures_Threshold_VerifyError(t *testing.T) {
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	signer := &CosmosMessageSignerRunnable{
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		chain: models.Chain{
			ChainID: "poktroll-different",
		},
		config: models.CosmosNetworkConfig{
			ChainID:           "poktroll-different",
			Bech32Prefix:      "pokt",
			MultisigThreshold: 1,
			MultisigAddress:   multisigAddr,
		},
	}

	mockClient.EXPECT().GetAccount(multisigAddr).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txConfig := util.NewTxConfig("pokt")
	txBuilder := txConfig.NewTxBuilder()

	signerData := authsigning.SignerData{
		ChainID:       "poktroll",
		AccountNumber: 1,
		Sequence:      1,
		PubKey:        signerKey.PubKey(),
		Address:       sdk.AccAddress(signerKey.PubKey().Address()).String(),
	}

	sigV2, _, err := util.SignWithPrivKey(context.Background(), signerData, txBuilder, signerKey, txConfig, 1)
	assert.NoError(t, err)

	err = txBuilder.SetSignatures(sigV2)
	assert.NoError(t, err)

	result := signer.ValidateSignaturesAndAddMultiSignatureToTxConfig("hash1", 1, txConfig, txBuilder)
	assert.False(t, result)

	mockClient.AssertExpectations(t)
}

func TestValidateSignatures_Threshold_AddSignatureError(t *testing.T) {
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	signer2Key := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signer2Key.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	signer := &CosmosMessageSignerRunnable{
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		chain: models.Chain{
			ChainID: "poktroll",
		},
		config: models.CosmosNetworkConfig{
			ChainID:           "poktroll",
			Bech32Prefix:      "pokt",
			MultisigThreshold: 1,
			MultisigAddress:   multisigAddr,
		},
	}

	mockClient.EXPECT().GetAccount(multisigAddr).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txConfig := util.NewTxConfig("pokt")
	txBuilder := txConfig.NewTxBuilder()

	signerData := authsigning.SignerData{
		ChainID:       "poktroll",
		AccountNumber: 1,
		Sequence:      1,
		PubKey:        signerKey.PubKey(),
		Address:       sdk.AccAddress(signerKey.PubKey().Address()).String(),
	}

	sigV2, _, err := util.SignWithPrivKey(context.Background(), signerData, txBuilder, signerKey, txConfig, 1)
	assert.NoError(t, err)

	err = txBuilder.SetSignatures(sigV2)
	assert.NoError(t, err)

	result := signer.ValidateSignaturesAndAddMultiSignatureToTxConfig("hash1", 1, txConfig, txBuilder)
	assert.False(t, result)

	mockClient.AssertExpectations(t)
}

func TestValidateSignatures_Threshold_AdaptableError(t *testing.T) {
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	signer := &CosmosMessageSignerRunnable{
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		chain: models.Chain{
			ChainID: "poktroll",
		},
		config: models.CosmosNetworkConfig{
			ChainID:           "poktroll",
			Bech32Prefix:      "pokt",
			MultisigThreshold: 1,
			MultisigAddress:   multisigAddr,
		},
	}

	mockClient.EXPECT().GetAccount(multisigAddr).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txConfig := util.NewTxConfig("pokt")
	txBuilder := clientMocks.NewMockTxBuilder(t)
	tx := clientMocks.NewMockTx(t)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{PubKey: signerKey.PubKey()}}, nil)
	txBuilder.EXPECT().GetTx().Return(tx)
	result := signer.ValidateSignaturesAndAddMultiSignatureToTxConfig("hash1", 1, txConfig, txBuilder)
	assert.False(t, result)

	mockClient.AssertExpectations(t)
}

func TestValidateSignatures_TwoThreshold(t *testing.T) {
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	signer2Key := secp256k1.GenPrivKey()
	signer3Key := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(2, []crypto.PubKey{signerKey.PubKey(), signer2Key.PubKey(), signer3Key.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	signer := &CosmosMessageSignerRunnable{
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		chain: models.Chain{
			ChainID: "poktroll",
		},
		config: models.CosmosNetworkConfig{
			ChainID:           "poktroll",
			Bech32Prefix:      "pokt",
			MultisigThreshold: 2,
			MultisigAddress:   multisigAddr,
		},
	}

	mockClient.EXPECT().GetAccount(multisigAddr).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txConfig := util.NewTxConfig("pokt")
	txBuilder := txConfig.NewTxBuilder()

	signerData := authsigning.SignerData{
		ChainID:       "poktroll",
		AccountNumber: 1,
		Sequence:      1,
		PubKey:        signerKey.PubKey(),
		Address:       sdk.AccAddress(signerKey.PubKey().Address()).String(),
	}

	sig1, _, err := util.SignWithPrivKey(context.Background(), signerData, txBuilder, signerKey, txConfig, 1)
	assert.NoError(t, err)

	sig3, _, err := util.SignWithPrivKey(context.Background(), signerData, txBuilder, signer3Key, txConfig, 1)
	assert.NoError(t, err)

	err = txBuilder.SetSignatures(sig1, sig3)
	assert.NoError(t, err)

	sigs, err := txBuilder.GetTx().GetSignaturesV2()
	assert.NoError(t, err)
	assert.Equal(t, sig1, sigs[0])
	assert.Equal(t, sig3, sigs[1])
	assert.Equal(t, 2, len(sigs))

	multisigSig := multisigtypes.NewMultisig(len(multisigPk.PubKeys))

	err = multisigtypes.AddSignatureV2(multisigSig, sig1, multisigPk.GetPubKeys())
	assert.NoError(t, err)
	err = multisigtypes.AddSignatureV2(multisigSig, sig3, multisigPk.GetPubKeys())
	assert.NoError(t, err)

	expectedSignature := signingtypes.SignatureV2{
		PubKey:   multisigPk,
		Data:     multisigSig,
		Sequence: 1,
	}

	result := signer.ValidateSignaturesAndAddMultiSignatureToTxConfig("hash1", 1, txConfig, txBuilder)
	assert.True(t, result)

	actualSigs, err := txBuilder.GetTx().GetSignaturesV2()
	assert.NoError(t, err)

	assert.Equal(t, expectedSignature, actualSigs[0])

	mockClient.AssertExpectations(t)
}

func TestFindMaxSequence(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockDB.EXPECT().LockReadSequences().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().FindMaxSequence(mock.Anything).Return(nil, nil)
	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	sequence, err := signer.FindMaxSequence()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), sequence)
}

func TestFindMaxSequence_LockError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockDB.EXPECT().LockReadSequences().Return("lock-id", assert.AnError)

	sequence, err := signer.FindMaxSequence()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), sequence)
}

func TestFindMaxSequence_FindError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockDB.EXPECT().LockReadSequences().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().FindMaxSequence(mock.Anything).Return(nil, assert.AnError)

	sequence, err := signer.FindMaxSequence()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), sequence)
}

func TestFindMaxSequence_AccountError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockDB.EXPECT().LockReadSequences().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().FindMaxSequence(mock.Anything).Return(nil, nil)
	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, assert.AnError)

	sequence, err := signer.FindMaxSequence()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), sequence)
}

func TestFindMaxSequence_AccountHigher(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockDB.EXPECT().LockReadSequences().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	seq := uint64(4)
	mockDB.EXPECT().FindMaxSequence(mock.Anything).Return(&seq, nil)
	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 5}, nil)

	sequence, err := signer.FindMaxSequence()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), sequence)
}

func TestFindMaxSequence_DBHigher(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockDB.EXPECT().LockReadSequences().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	seq := uint64(4)
	mockDB.EXPECT().FindMaxSequence(mock.Anything).Return(&seq, nil)
	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 3}, nil)

	sequence, err := signer.FindMaxSequence()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), sequence)
}

func TestBroadcastMessage_WrapError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
	}

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, assert.AnError
	}

	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
	}()

	result := signer.BroadcastMessage(message)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.False(t, result)
}

func TestBroadcastMessage(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	tx := clientMocks.NewMockTx(t)
	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)

	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	})

	txConfig.EXPECT().TxEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx as bytes"), nil
	})

	mockClient.EXPECT().BroadcastTx([]byte("encoded tx as bytes")).Return("0x0102", nil)
	expectedUpdate := bson.M{
		"status":           models.MessageStatusBroadcasted,
		"transaction_hash": "0x0102",
		"transaction_body": "encoded tx",
	}

	mockDB.EXPECT().UpdateMessage(message.ID, expectedUpdate).Return(nil)

	result := signer.BroadcastMessage(message)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.True(t, result)
}

func TestBroadcastMessage_Invalid(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
	}()

	tx := clientMocks.NewMockTx(t)
	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(assert.AnError)

	expectedUpdate := bson.M{
		"status":           models.MessageStatusPending,
		"transaction":      nil,
		"transaction_hash": "",
		"transaction_body": "",
		"signatures":       []models.Signature{},
	}

	mockDB.EXPECT().UpdateMessage(message.ID, expectedUpdate).Return(nil)

	result := signer.BroadcastMessage(message)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.True(t, result)
}

func TestBroadcastMessage_JsonEncoderError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
	}()

	tx := clientMocks.NewMockTx(t)
	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return nil, assert.AnError
	})

	result := signer.BroadcastMessage(message)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.False(t, result)
}

func TestBroadcastMessage_EncoderError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
	}()

	tx := clientMocks.NewMockTx(t)
	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	})

	txConfig.EXPECT().TxEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx as bytes"), assert.AnError
	})

	result := signer.BroadcastMessage(message)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.False(t, result)
}

func TestBroadcastMessage_BroadcastError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
	}()

	tx := clientMocks.NewMockTx(t)
	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	})

	txConfig.EXPECT().TxEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx as bytes"), nil
	})

	mockClient.EXPECT().BroadcastTx([]byte("encoded tx as bytes")).Return("0x0102", assert.AnError)

	result := signer.BroadcastMessage(message)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateEthereumTxAndBroadcastMessage_ValidateError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             "0xinvalid",
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	result := signer.ValidateEthereumTxAndBroadcastMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateEthereumTxAndBroadcastMessage_Pending(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(91), nil)
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	result := signer.ValidateEthereumTxAndBroadcastMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateEthereumTxAndBroadcastMessage_FailedTx(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusFailed,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)

	mockDB.EXPECT().UpdateMessage(message.ID, bson.M{"status": models.MessageStatusInvalid}).Return(nil)

	result := signer.ValidateEthereumTxAndBroadcastMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.True(t, result)
}

func TestValidateEthereumTxAndBroadcastMessage_LockError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	mockDB.EXPECT().LockWriteMessage(mock.Anything).Return("lock-id", assert.AnError)

	result := signer.ValidateEthereumTxAndBroadcastMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateEthereumTxAndBroadcastMessage(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)
	tx := clientMocks.NewMockTx(t)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)
	txBuilder.EXPECT().GetTx().Return(tx)

	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	mockClient.EXPECT().GetAccount(multisigAddr).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	mockDB.EXPECT().LockWriteMessage(mock.Anything).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	})

	txConfig.EXPECT().TxEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx as bytes"), nil
	})

	mockClient.EXPECT().BroadcastTx([]byte("encoded tx as bytes")).Return("0x0102", nil)
	expectedUpdate := bson.M{
		"status":           models.MessageStatusBroadcasted,
		"transaction_hash": "0x0102",
		"transaction_body": "encoded tx",
	}

	mockDB.EXPECT().UpdateMessage(message.ID, expectedUpdate).Return(nil)

	result := signer.ValidateEthereumTxAndBroadcastMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.True(t, result)
}

func TestBroadcastMessages(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)
	tx := clientMocks.NewMockTx(t)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)
	txBuilder.EXPECT().GetTx().Return(tx)

	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	mockClient.EXPECT().GetAccount(multisigAddr).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	mockDB.EXPECT().LockWriteMessage(mock.Anything).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	})

	txConfig.EXPECT().TxEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx as bytes"), nil
	})

	mockClient.EXPECT().BroadcastTx([]byte("encoded tx as bytes")).Return("0x0102", nil)
	expectedUpdate := bson.M{
		"status":           models.MessageStatusBroadcasted,
		"transaction_hash": "0x0102",
		"transaction_body": "encoded tx",
	}

	mockDB.EXPECT().UpdateMessage(message.ID, expectedUpdate).Return(nil)

	mockDB.EXPECT().GetSignedMessages(mock.Anything).Return([]models.Message{*message}, nil)

	result := signer.BroadcastMessages()

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.True(t, result)
}

func TestBroadcastMessages_GetSignedMessagesError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().GetSignedMessages(mock.Anything).Return(nil, assert.AnError)

	result := signer.BroadcastMessages()

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestBroadcastRefund_WrapError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Recipient:             "",
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, assert.AnError
	}

	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
	}()

	result := signer.BroadcastRefund(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.False(t, result)
}

func TestBroadcastRefund(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	tx := clientMocks.NewMockTx(t)
	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	})

	txConfig.EXPECT().TxEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx as bytes"), nil
	})

	mockClient.EXPECT().BroadcastTx([]byte("encoded tx as bytes")).Return("0x0102", nil)
	expectedUpdate := bson.M{
		"status":           models.RefundStatusBroadcasted,
		"transaction_hash": "0x0102",
		"transaction_body": "encoded tx",
	}

	mockDB.EXPECT().UpdateRefund(refund.ID, expectedUpdate).Return(nil)

	result := signer.BroadcastRefund(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.True(t, result)
}

func TestBroadcastRefund_Invalid(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
	}()

	tx := clientMocks.NewMockTx(t)
	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(assert.AnError)

	expectedUpdate := bson.M{
		"status":           models.RefundStatusPending,
		"transaction":      nil,
		"transaction_hash": "",
		"transaction_body": "",
		"signatures":       []models.Signature{},
	}

	mockDB.EXPECT().UpdateRefund(refund.ID, expectedUpdate).Return(nil)

	result := signer.BroadcastRefund(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.True(t, result)
}

func TestBroadcastRefund_JsonEncoderError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
	}()

	tx := clientMocks.NewMockTx(t)
	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return nil, assert.AnError
	})

	result := signer.BroadcastRefund(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.False(t, result)
}

func TestBroadcastRefund_EncoderError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
	}()

	tx := clientMocks.NewMockTx(t)
	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	})

	txConfig.EXPECT().TxEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx as bytes"), assert.AnError
	})

	result := signer.BroadcastRefund(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.False(t, result)
}

func TestBroadcastRefund_BroadcastError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
	}()

	tx := clientMocks.NewMockTx(t)
	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	})

	txConfig.EXPECT().TxEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx as bytes"), nil
	})

	mockClient.EXPECT().BroadcastTx([]byte("encoded tx as bytes")).Return("0x0102", assert.AnError)

	result := signer.BroadcastRefund(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateRefund_EmptyBody(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	result := signer.ValidateRefund(refund, recipientAddr.Bytes(), amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestValidateRefund_InvalidAddress(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             "recipient",
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	result := signer.ValidateRefund(refund, recipientAddr.Bytes(), amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateRefund_AddressError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	spenderAddr := ethcommon.BytesToAddress([]byte("spender"))

	result := signer.ValidateRefund(refund, spenderAddr.Bytes(), amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateRefund_InvalidAmount(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "invalid",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	result := signer.ValidateRefund(refund, recipientAddr[:], amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateRefund_AmountError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("1000upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	result := signer.ValidateRefund(refund, recipientAddr[:], amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateRefund_WithBody(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddrBech32, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	recipientAddrBech32, _ := common.Bech32FromBytes("pokt", recipientAddr.Bytes())
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
		TransactionBody:       "txBody",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddrBech32,
		},
	}

	mockTx := clientMocks.NewMockTx(t)
	utilParseTxBody = func(string, string) (sdk.Tx, error) {
		return mockTx, nil
	}

	validMsg := &banktypes.MsgSend{FromAddress: multisigAddrBech32, ToAddress: recipientAddrBech32, Amount: sdk.NewCoins(amount)}

	mockTx.EXPECT().GetMsgs().Return([]sdk.Msg{validMsg})

	result := signer.ValidateRefund(refund, recipientAddr.Bytes(), amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestValidateRefund_WithBody_ParseError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddrBech32, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
		TransactionBody:       "txBody",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddrBech32,
		},
	}

	utilParseTxBody = func(string, string) (sdk.Tx, error) {
		return nil, assert.AnError
	}
	result := signer.ValidateRefund(refund, recipientAddr.Bytes(), amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateRefund_WithBody_InvalidMsg(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddrBech32, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	recipientAddrBech32, _ := common.Bech32FromBytes("pokt", recipientAddr.Bytes())
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
		TransactionBody:       "txBody",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddrBech32,
		},
	}

	mockTx := clientMocks.NewMockTx(t)
	utilParseTxBody = func(string, string) (sdk.Tx, error) {
		return mockTx, nil
	}

	validMsg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{Address: multisigAddrBech32, Coins: sdk.NewCoins(amount)},
		},
		Outputs: []banktypes.Output{
			{Address: recipientAddrBech32, Coins: sdk.NewCoins(amount)},
		},
	}

	mockTx.EXPECT().GetMsgs().Return([]sdk.Msg{validMsg})

	result := signer.ValidateRefund(refund, recipientAddr.Bytes(), amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateRefund_WithBody_InvalidAmount(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddrBech32, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	recipientAddrBech32, _ := common.Bech32FromBytes("pokt", recipientAddr.Bytes())
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
		TransactionBody:       "txBody",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddrBech32,
		},
	}

	mockTx := clientMocks.NewMockTx(t)
	utilParseTxBody = func(string, string) (sdk.Tx, error) {
		return mockTx, nil
	}

	amount2, _ := sdk.ParseCoinNormalized("100atom")

	validMsg := &banktypes.MsgSend{FromAddress: multisigAddrBech32, ToAddress: recipientAddrBech32, Amount: sdk.NewCoins(amount, amount2)}

	mockTx.EXPECT().GetMsgs().Return([]sdk.Msg{validMsg})

	result := signer.ValidateRefund(refund, recipientAddr.Bytes(), amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateRefund_WithBody_DifferentAmount(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddrBech32, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	recipientAddrBech32, _ := common.Bech32FromBytes("pokt", recipientAddr.Bytes())
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
		TransactionBody:       "txBody",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddrBech32,
		},
	}

	mockTx := clientMocks.NewMockTx(t)
	utilParseTxBody = func(string, string) (sdk.Tx, error) {
		return mockTx, nil
	}

	validMsg := &banktypes.MsgSend{FromAddress: multisigAddrBech32, ToAddress: recipientAddrBech32, Amount: sdk.NewCoins(amount.Add(amount))}

	mockTx.EXPECT().GetMsgs().Return([]sdk.Msg{validMsg})

	result := signer.ValidateRefund(refund, recipientAddr.Bytes(), amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateRefund_WithBody_FromAddressError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddrBech32, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	recipientAddrBech32, _ := common.Bech32FromBytes("pokt", recipientAddr.Bytes())
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
		TransactionBody:       "txBody",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddrBech32,
		},
	}

	mockTx := clientMocks.NewMockTx(t)
	utilParseTxBody = func(string, string) (sdk.Tx, error) {
		return mockTx, nil
	}

	validMsg := &banktypes.MsgSend{FromAddress: "multisigAddrBech32", ToAddress: recipientAddrBech32, Amount: sdk.NewCoins(amount)}

	mockTx.EXPECT().GetMsgs().Return([]sdk.Msg{validMsg})

	result := signer.ValidateRefund(refund, recipientAddr.Bytes(), amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateRefund_WithBody_FromAddressDifferentError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddrBech32, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	recipientAddrBech32, _ := common.Bech32FromBytes("pokt", recipientAddr.Bytes())
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
		TransactionBody:       "txBody",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddrBech32,
		},
	}

	mockTx := clientMocks.NewMockTx(t)
	utilParseTxBody = func(string, string) (sdk.Tx, error) {
		return mockTx, nil
	}

	validMsg := &banktypes.MsgSend{FromAddress: recipientAddrBech32, ToAddress: recipientAddrBech32, Amount: sdk.NewCoins(amount)}

	mockTx.EXPECT().GetMsgs().Return([]sdk.Msg{validMsg})

	result := signer.ValidateRefund(refund, recipientAddr.Bytes(), amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateRefund_WithBody_ToAddressError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddrBech32, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
		TransactionBody:       "txBody",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddrBech32,
		},
	}

	mockTx := clientMocks.NewMockTx(t)
	utilParseTxBody = func(string, string) (sdk.Tx, error) {
		return mockTx, nil
	}

	validMsg := &banktypes.MsgSend{FromAddress: multisigAddrBech32, ToAddress: "recipientAddrBech32", Amount: sdk.NewCoins(amount)}

	mockTx.EXPECT().GetMsgs().Return([]sdk.Msg{validMsg})

	result := signer.ValidateRefund(refund, recipientAddr.Bytes(), amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateRefund_WithBody_ToAddressDifferentError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddrBech32, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := &models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
		TransactionBody:       "txBody",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddrBech32,
		},
	}

	mockTx := clientMocks.NewMockTx(t)
	utilParseTxBody = func(string, string) (sdk.Tx, error) {
		return mockTx, nil
	}

	validMsg := &banktypes.MsgSend{FromAddress: multisigAddrBech32, ToAddress: multisigAddrBech32, Amount: sdk.NewCoins(amount)}

	mockTx.EXPECT().GetMsgs().Return([]sdk.Msg{validMsg})

	result := signer.ValidateRefund(refund, recipientAddr.Bytes(), amount)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateCosmosTx_ClientError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(nil, assert.AnError)

	result := signer.ValidateCosmosTx(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateCosmosTx_ValidateError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return nil, assert.AnError
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().UpdateRefund(mock.Anything, mock.Anything).Return(nil)

	result := signer.ValidateCosmosTx(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateCosmosTx_Pending(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			TxStatus:      models.TransactionStatusPending,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	result := signer.ValidateCosmosTx(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateCosmosTx_NoRefund(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			NeedsRefund:   false,
			TxStatus:      models.TransactionStatusConfirmed,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().UpdateRefund(mock.Anything, mock.Anything).Return(nil)

	result := signer.ValidateCosmosTx(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateCosmosTx_Failed(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			NeedsRefund:   true,
			TxStatus:      models.TransactionStatusFailed,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().UpdateRefund(mock.Anything, mock.Anything).Return(nil)

	result := signer.ValidateCosmosTx(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateCosmosTx_Successful(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			NeedsRefund:   true,
			TxStatus:      models.TransactionStatusConfirmed,
			SenderAddress: recipientAddr[:],
			Amount:        amount,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	result := signer.ValidateCosmosTx(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestValidateCosmosTx_Invalid(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			NeedsRefund:   true,
			TxStatus:      models.TransactionStatusConfirmed,
			Amount:        amount,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().UpdateRefund(mock.Anything, mock.Anything).Return(nil)

	result := signer.ValidateCosmosTx(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateCosmosTxAndBroadcastRefund_Invalid(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			NeedsRefund:   true,
			TxStatus:      models.TransactionStatusConfirmed,
			Amount:        amount,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().UpdateRefund(mock.Anything, mock.Anything).Return(nil)

	result := signer.ValidateCosmosTxAndBroadcastRefund(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateCosmosTxAndBroadcastRefund_LockError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			NeedsRefund:   true,
			TxStatus:      models.TransactionStatusConfirmed,
			SenderAddress: recipientAddr[:],
			Amount:        amount,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().LockWriteRefund(&refund).Return("lock-id", assert.AnError)

	result := signer.ValidateCosmosTxAndBroadcastRefund(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateCosmosTxAndBroadcastRefund(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			NeedsRefund:   true,
			TxStatus:      models.TransactionStatusConfirmed,
			SenderAddress: recipientAddr[:],
			Amount:        amount,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().LockWriteRefund(&refund).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	tx := clientMocks.NewMockTx(t)
	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	})

	txConfig.EXPECT().TxEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx as bytes"), nil
	})

	mockClient.EXPECT().BroadcastTx([]byte("encoded tx as bytes")).Return("0x0102", nil)

	mockDB.EXPECT().UpdateRefund(mock.Anything, mock.Anything).Return(nil)

	result := signer.ValidateCosmosTxAndBroadcastRefund(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestBroadcastRefunds(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			NeedsRefund:   true,
			TxStatus:      models.TransactionStatusConfirmed,
			SenderAddress: recipientAddr[:],
			Amount:        amount,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().LockWriteRefund(&refund).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
		return txBuilder, txConfig, nil
	}

	tx := clientMocks.NewMockTx(t)
	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{{}}, nil)
	utilValidateSignature = func(models.CosmosNetworkConfig, *signingtypes.SignatureV2, uint64, uint64, client.TxConfig, client.TxBuilder,
	) error {
		return nil
	}
	multisigtypesAddSignatureV2 = func(*signingtypes.MultiSignatureData, signingtypes.SignatureV2, []crypto.PubKey) error {
		return nil
	}
	defer func() {
		utilWrapTxBuilder = util.WrapTxBuilder
		utilValidateSignature = util.ValidateSignature
		multisigtypesAddSignatureV2 = multisigtypes.AddSignatureV2
	}()

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	})

	txConfig.EXPECT().TxEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx as bytes"), nil
	})

	mockClient.EXPECT().BroadcastTx([]byte("encoded tx as bytes")).Return("0x0102", nil)

	mockDB.EXPECT().UpdateRefund(mock.Anything, mock.Anything).Return(nil)

	mockDB.EXPECT().GetSignedRefunds().Return([]models.Refund{refund}, nil)

	result := signer.BroadcastRefunds()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestBroadcastRefunds_Error(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockDB.EXPECT().GetSignedRefunds().Return(nil, assert.AnError)

	result := signer.BroadcastRefunds()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignRefunds_Error(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockDB.EXPECT().GetPendingRefunds(mock.Anything).Return(nil, assert.AnError)

	result := signer.SignRefunds()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateCosmosTxAndSignRefund_Invalid(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			NeedsRefund:   true,
			TxStatus:      models.TransactionStatusConfirmed,
			Amount:        amount,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().UpdateRefund(mock.Anything, mock.Anything).Return(nil)

	result := signer.ValidateCosmosTxAndSignRefund(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateCosmosTxAndSignRefund_LockError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			NeedsRefund:   true,
			TxStatus:      models.TransactionStatusConfirmed,
			SenderAddress: recipientAddr[:],
			Amount:        amount,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().LockWriteRefund(&refund).Return("lock-id", assert.AnError)

	result := signer.ValidateCosmosTxAndSignRefund(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateCosmosTxAndSignRefund(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			NeedsRefund:   true,
			TxStatus:      models.TransactionStatusConfirmed,
			SenderAddress: recipientAddr[:],
			Amount:        amount,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().LockWriteRefund(&refund).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, nil
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	mockDB.EXPECT().UpdateRefund(refund.ID, mock.Anything).Return(nil)

	result := signer.ValidateCosmosTxAndSignRefund(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignRefunds(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))
	amount, _ := sdk.ParseCoinNormalized("100upokt")

	refund := models.Refund{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
		Recipient:             recipientAddr.Hex(),
		Amount:                "100",
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetTx(refund.OriginTransactionHash).Return(&sdk.TxResponse{}, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		result := &util.ValidateTxResult{
			Confirmations: 0,
			NeedsRefund:   true,
			TxStatus:      models.TransactionStatusConfirmed,
			SenderAddress: recipientAddr[:],
			Amount:        amount,
		}
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().LockWriteRefund(&refund).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	oldCosmosSignTx := CosmosSignTx
	CosmosSignTx = func(
		signerKey crypto.PrivKey,
		config models.CosmosNetworkConfig,
		client cosmos.CosmosClient,
		sequence uint64,
		signatures []models.Signature,
		transactionBody string,
		toAddress []byte,
		amount sdk.Coin,
		memo string,
	) (string, []models.Signature, error) {
		return "encoded tx", []models.Signature{{}}, nil
	}
	defer func() { CosmosSignTx = oldCosmosSignTx }()

	mockDB.EXPECT().UpdateRefund(refund.ID, mock.Anything).Return(nil)

	mockDB.EXPECT().GetPendingRefunds(mock.Anything).Return([]models.Refund{refund}, nil)

	result := signer.SignRefunds()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignerRun(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []crypto.PubKey{signerKey.PubKey()})
	multisigAddr, _ := common.Bech32FromBytes("pokt", multisigPk.Address().Bytes())

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		signerKey:  signerKey,
		multisigPk: multisigPk,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: multisigAddr,
		},
	}

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)
	mockDB.EXPECT().GetPendingRefunds(mock.Anything).Return([]models.Refund{}, nil)
	mockDB.EXPECT().GetSignedRefunds().Return([]models.Refund{}, nil)
	mockDB.EXPECT().GetPendingMessages(mock.Anything, mock.Anything).Return([]models.Message{}, nil)
	mockDB.EXPECT().GetSignedMessages(mock.Anything).Return([]models.Message{}, nil)

	signer.Run()
}

func TestNewMessageSigner(t *testing.T) {
	mnemonic := "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	config := models.CosmosNetworkConfig{
		StartBlockHeight:   1,
		Confirmations:      1,
		RPCURL:             "http://localhost:36657",
		GRPCEnabled:        true,
		GRPCHost:           "localhost",
		GRPCPort:           9090,
		TimeoutMS:          1000,
		ChainID:            "poktroll",
		ChainName:          "Poktroll",
		TxFee:              1000,
		Bech32Prefix:       "pokt",
		CoinDenom:          "upokt",
		MultisigAddress:    "pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar",
		MultisigPublicKeys: []string{"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9", "02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2", "02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6df"},
		MultisigThreshold:  2,
		MessageMonitor: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageSigner: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageRelayer: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
	}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = []byte("mintControllerAddress")

	ethNetworks := []models.EthereumNetworkConfig{
		{
			StartBlockHeight:      1,
			Confirmations:         1,
			RPCURL:                "http://localhost:8545",
			TimeoutMS:             1000,
			ChainID:               1,
			ChainName:             "Ethereum",
			MailboxAddress:        "0x0000000000000000000000000000000000000000",
			MintControllerAddress: "0x0000000000000000000000000000000000000000",
			OmniTokenAddress:      "0x0000000000000000000000000000000000000000",
			WarpISMAddress:        "0x0000000000000000000000000000000000000000",
			OracleAddresses:       []string{"0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb", "0x958d1F55E14Cba24a077b9634F16f83565fc9411", "0x4c672Edd2ec8eac8f0F1709f33de9A2E786e6912"},
			MessageMonitor: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageSigner: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageRelayer: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
		},
	}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := dbMocks.NewMockDB(t)

	// Mocking client methods
	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	originalNewDB := dbNewDB
	defer func() { dbNewDB = originalNewDB }()
	dbNewDB = func() db.DB {
		return mockDB
	}

	originalCosmosNewClient := cosmosNewClient
	defer func() { cosmosNewClient = originalCosmosNewClient }()
	cosmosNewClient = func(config models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
		return mockClient, nil
	}

	originEthNewClient := ethNewClient
	defer func() { ethNewClient = originEthNewClient }()
	ethNewClient = func(config models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		mockEthClient := ethMocks.NewMockEthereumClient(t)
		mockEthClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(config.ChainID)})
		mockEthClient.EXPECT().GetClient().Return(nil)

		return mockEthClient, nil
	}

	originalEthNewMailboxContract := ethNewMailboxContract
	defer func() { ethNewMailboxContract = originalEthNewMailboxContract }()
	ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
		return nil, nil
	}

	runnable := NewMessageSigner(mnemonic, config, mintControllerMap, ethNetworks)

	assert.NotNil(t, runnable)
	monitor, ok := runnable.(*CosmosMessageSignerRunnable)
	assert.True(t, ok)

	assert.Equal(t, uint64(100), monitor.currentBlockHeight)
	assert.Equal(t, config, monitor.config)
	assert.Equal(t, util.ParseChain(config), monitor.chain)
	assert.Equal(t, mintControllerMap, monitor.mintControllerMap)
	assert.NotNil(t, monitor.client)
	assert.NotNil(t, monitor.logger)
	assert.NotNil(t, monitor.db)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestNewMessageSigner_Disabled(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	mnemonic := "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	config := models.CosmosNetworkConfig{
		StartBlockHeight:   1,
		Confirmations:      1,
		RPCURL:             "http://localhost:36657",
		GRPCEnabled:        true,
		GRPCHost:           "localhost",
		GRPCPort:           9090,
		TimeoutMS:          1000,
		ChainID:            "poktroll",
		ChainName:          "Poktroll",
		TxFee:              1000,
		Bech32Prefix:       "pokt",
		CoinDenom:          "upokt",
		MultisigAddress:    "pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar",
		MultisigPublicKeys: []string{"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9", "02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2", "02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6df"},
		MultisigThreshold:  2,
		MessageMonitor: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageSigner: models.ServiceConfig{
			Enabled:    false,
			IntervalMS: 1000,
		},
		MessageRelayer: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
	}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = []byte("mintControllerAddress")

	ethNetworks := []models.EthereumNetworkConfig{
		{
			StartBlockHeight:      1,
			Confirmations:         1,
			RPCURL:                "http://localhost:8545",
			TimeoutMS:             1000,
			ChainID:               1,
			ChainName:             "Ethereum",
			MailboxAddress:        "0x0000000000000000000000000000000000000000",
			MintControllerAddress: "0x0000000000000000000000000000000000000000",
			OmniTokenAddress:      "0x0000000000000000000000000000000000000000",
			WarpISMAddress:        "0x0000000000000000000000000000000000000000",
			OracleAddresses:       []string{"0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb", "0x958d1F55E14Cba24a077b9634F16f83565fc9411", "0x4c672Edd2ec8eac8f0F1709f33de9A2E786e6912"},
			MessageMonitor: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageSigner: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageRelayer: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
		},
	}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := dbMocks.NewMockDB(t)

	// Mocking client methods
	// mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	originalNewDB := dbNewDB
	defer func() { dbNewDB = originalNewDB }()
	dbNewDB = func() db.DB {
		return mockDB
	}

	originalCosmosNewClient := cosmosNewClient
	defer func() { cosmosNewClient = originalCosmosNewClient }()
	cosmosNewClient = func(config models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
		return mockClient, nil
	}

	originEthNewClient := ethNewClient
	defer func() { ethNewClient = originEthNewClient }()
	ethNewClient = func(config models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		mockEthClient := ethMocks.NewMockEthereumClient(t)
		// mockEthClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(config.ChainID)})
		// mockEthClient.EXPECT().GetClient().Return(nil)

		return mockEthClient, nil
	}

	originalEthNewMailboxContract := ethNewMailboxContract
	defer func() { ethNewMailboxContract = originalEthNewMailboxContract }()
	ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
		return nil, nil
	}

	assert.Panics(t, func() {
		NewMessageSigner(mnemonic, config, mintControllerMap, ethNetworks)
	})
}

func TestNewMessageSigner_InvalidPublicKey(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	mnemonic := "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	config := models.CosmosNetworkConfig{
		StartBlockHeight:   1,
		Confirmations:      1,
		RPCURL:             "http://localhost:36657",
		GRPCEnabled:        true,
		GRPCHost:           "localhost",
		GRPCPort:           9090,
		TimeoutMS:          1000,
		ChainID:            "poktroll",
		ChainName:          "Poktroll",
		TxFee:              1000,
		Bech32Prefix:       "pokt",
		CoinDenom:          "upokt",
		MultisigAddress:    "pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar",
		MultisigPublicKeys: []string{"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4", "02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2", "02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6df"},
		MultisigThreshold:  2,
		MessageMonitor: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageSigner: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageRelayer: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
	}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = []byte("mintControllerAddress")

	ethNetworks := []models.EthereumNetworkConfig{
		{
			StartBlockHeight:      1,
			Confirmations:         1,
			RPCURL:                "http://localhost:8545",
			TimeoutMS:             1000,
			ChainID:               1,
			ChainName:             "Ethereum",
			MailboxAddress:        "0x0000000000000000000000000000000000000000",
			MintControllerAddress: "0x0000000000000000000000000000000000000000",
			OmniTokenAddress:      "0x0000000000000000000000000000000000000000",
			WarpISMAddress:        "0x0000000000000000000000000000000000000000",
			OracleAddresses:       []string{"0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb", "0x958d1F55E14Cba24a077b9634F16f83565fc9411", "0x4c672Edd2ec8eac8f0F1709f33de9A2E786e6912"},
			MessageMonitor: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageSigner: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageRelayer: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
		},
	}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := dbMocks.NewMockDB(t)

	// Mocking client methods
	// mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	originalNewDB := dbNewDB
	defer func() { dbNewDB = originalNewDB }()
	dbNewDB = func() db.DB {
		return mockDB
	}

	originalCosmosNewClient := cosmosNewClient
	defer func() { cosmosNewClient = originalCosmosNewClient }()
	cosmosNewClient = func(config models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
		return mockClient, nil
	}

	originEthNewClient := ethNewClient
	defer func() { ethNewClient = originEthNewClient }()
	ethNewClient = func(config models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		mockEthClient := ethMocks.NewMockEthereumClient(t)
		// mockEthClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(config.ChainID)})
		// mockEthClient.EXPECT().GetClient().Return(nil)

		return mockEthClient, nil
	}

	originalEthNewMailboxContract := ethNewMailboxContract
	defer func() { ethNewMailboxContract = originalEthNewMailboxContract }()
	ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
		return nil, nil
	}

	assert.Panics(t, func() {
		NewMessageSigner(mnemonic, config, mintControllerMap, ethNetworks)
	})
}

func TestNewMessageSigner_InvalidMultisigAddress(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	mnemonic := "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	config := models.CosmosNetworkConfig{
		StartBlockHeight:   1,
		Confirmations:      1,
		RPCURL:             "http://localhost:36657",
		GRPCEnabled:        true,
		GRPCHost:           "localhost",
		GRPCPort:           9090,
		TimeoutMS:          1000,
		ChainID:            "poktroll",
		ChainName:          "Poktroll",
		TxFee:              1000,
		Bech32Prefix:       "pokt",
		CoinDenom:          "upokt",
		MultisigAddress:    "pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2",
		MultisigPublicKeys: []string{"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9", "02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2", "02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6df"},
		MultisigThreshold:  2,
		MessageMonitor: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageSigner: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageRelayer: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
	}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = []byte("mintControllerAddress")

	ethNetworks := []models.EthereumNetworkConfig{
		{
			StartBlockHeight:      1,
			Confirmations:         1,
			RPCURL:                "http://localhost:8545",
			TimeoutMS:             1000,
			ChainID:               1,
			ChainName:             "Ethereum",
			MailboxAddress:        "0x0000000000000000000000000000000000000000",
			MintControllerAddress: "0x0000000000000000000000000000000000000000",
			OmniTokenAddress:      "0x0000000000000000000000000000000000000000",
			WarpISMAddress:        "0x0000000000000000000000000000000000000000",
			OracleAddresses:       []string{"0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb", "0x958d1F55E14Cba24a077b9634F16f83565fc9411", "0x4c672Edd2ec8eac8f0F1709f33de9A2E786e6912"},
			MessageMonitor: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageSigner: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageRelayer: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
		},
	}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := dbMocks.NewMockDB(t)

	// Mocking client methods
	// mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	originalNewDB := dbNewDB
	defer func() { dbNewDB = originalNewDB }()
	dbNewDB = func() db.DB {
		return mockDB
	}

	originalCosmosNewClient := cosmosNewClient
	defer func() { cosmosNewClient = originalCosmosNewClient }()
	cosmosNewClient = func(config models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
		return mockClient, nil
	}

	originEthNewClient := ethNewClient
	defer func() { ethNewClient = originEthNewClient }()
	ethNewClient = func(config models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		mockEthClient := ethMocks.NewMockEthereumClient(t)
		// mockEthClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(config.ChainID)})
		// mockEthClient.EXPECT().GetClient().Return(nil)

		return mockEthClient, nil
	}

	originalEthNewMailboxContract := ethNewMailboxContract
	defer func() { ethNewMailboxContract = originalEthNewMailboxContract }()
	ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
		return nil, nil
	}

	assert.Panics(t, func() {
		NewMessageSigner(mnemonic, config, mintControllerMap, ethNetworks)
	})
}

func TestNewMessageSigner_ClientError(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	mnemonic := "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	config := models.CosmosNetworkConfig{
		StartBlockHeight:   1,
		Confirmations:      1,
		RPCURL:             "http://localhost:36657",
		GRPCEnabled:        true,
		GRPCHost:           "localhost",
		GRPCPort:           9090,
		TimeoutMS:          1000,
		ChainID:            "poktroll",
		ChainName:          "Poktroll",
		TxFee:              1000,
		Bech32Prefix:       "pokt",
		CoinDenom:          "upokt",
		MultisigAddress:    "pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar",
		MultisigPublicKeys: []string{"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9", "02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2", "02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6df"},
		MultisigThreshold:  2,
		MessageMonitor: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageSigner: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageRelayer: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
	}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = []byte("mintControllerAddress")

	ethNetworks := []models.EthereumNetworkConfig{
		{
			StartBlockHeight:      1,
			Confirmations:         1,
			RPCURL:                "http://localhost:8545",
			TimeoutMS:             1000,
			ChainID:               1,
			ChainName:             "Ethereum",
			MailboxAddress:        "0x0000000000000000000000000000000000000000",
			MintControllerAddress: "0x0000000000000000000000000000000000000000",
			OmniTokenAddress:      "0x0000000000000000000000000000000000000000",
			WarpISMAddress:        "0x0000000000000000000000000000000000000000",
			OracleAddresses:       []string{"0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb", "0x958d1F55E14Cba24a077b9634F16f83565fc9411", "0x4c672Edd2ec8eac8f0F1709f33de9A2E786e6912"},
			MessageMonitor: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageSigner: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageRelayer: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
		},
	}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := dbMocks.NewMockDB(t)

	// Mocking client methods
	// mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	originalNewDB := dbNewDB
	defer func() { dbNewDB = originalNewDB }()
	dbNewDB = func() db.DB {
		return mockDB
	}

	originalCosmosNewClient := cosmosNewClient
	defer func() { cosmosNewClient = originalCosmosNewClient }()
	cosmosNewClient = func(config models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
		return mockClient, assert.AnError
	}

	originEthNewClient := ethNewClient
	defer func() { ethNewClient = originEthNewClient }()
	ethNewClient = func(config models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		mockEthClient := ethMocks.NewMockEthereumClient(t)
		// mockEthClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(config.ChainID)})
		// mockEthClient.EXPECT().GetClient().Return(nil)

		return mockEthClient, nil
	}

	originalEthNewMailboxContract := ethNewMailboxContract
	defer func() { ethNewMailboxContract = originalEthNewMailboxContract }()
	ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
		return nil, nil
	}

	assert.Panics(t, func() {
		NewMessageSigner(mnemonic, config, mintControllerMap, ethNetworks)
	})
}

func TestNewMessageSigner_MnemonicError(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	// mnemonic := "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	config := models.CosmosNetworkConfig{
		StartBlockHeight:   1,
		Confirmations:      1,
		RPCURL:             "http://localhost:36657",
		GRPCEnabled:        true,
		GRPCHost:           "localhost",
		GRPCPort:           9090,
		TimeoutMS:          1000,
		ChainID:            "poktroll",
		ChainName:          "Poktroll",
		TxFee:              1000,
		Bech32Prefix:       "pokt",
		CoinDenom:          "upokt",
		MultisigAddress:    "pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar",
		MultisigPublicKeys: []string{"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9", "02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2", "02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6df"},
		MultisigThreshold:  2,
		MessageMonitor: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageSigner: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageRelayer: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
	}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = []byte("mintControllerAddress")

	ethNetworks := []models.EthereumNetworkConfig{
		{
			StartBlockHeight:      1,
			Confirmations:         1,
			RPCURL:                "http://localhost:8545",
			TimeoutMS:             1000,
			ChainID:               1,
			ChainName:             "Ethereum",
			MailboxAddress:        "0x0000000000000000000000000000000000000000",
			MintControllerAddress: "0x0000000000000000000000000000000000000000",
			OmniTokenAddress:      "0x0000000000000000000000000000000000000000",
			WarpISMAddress:        "0x0000000000000000000000000000000000000000",
			OracleAddresses:       []string{"0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb", "0x958d1F55E14Cba24a077b9634F16f83565fc9411", "0x4c672Edd2ec8eac8f0F1709f33de9A2E786e6912"},
			MessageMonitor: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageSigner: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageRelayer: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
		},
	}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := dbMocks.NewMockDB(t)

	// Mocking client methods
	// mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	originalNewDB := dbNewDB
	defer func() { dbNewDB = originalNewDB }()
	dbNewDB = func() db.DB {
		return mockDB
	}

	originalCosmosNewClient := cosmosNewClient
	defer func() { cosmosNewClient = originalCosmosNewClient }()
	cosmosNewClient = func(config models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
		return mockClient, nil
	}

	originEthNewClient := ethNewClient
	defer func() { ethNewClient = originEthNewClient }()
	ethNewClient = func(config models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		mockEthClient := ethMocks.NewMockEthereumClient(t)
		// mockEthClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(config.ChainID)})
		// mockEthClient.EXPECT().GetClient().Return(nil)

		return mockEthClient, nil
	}

	originalEthNewMailboxContract := ethNewMailboxContract
	defer func() { ethNewMailboxContract = originalEthNewMailboxContract }()
	ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
		return nil, nil
	}

	assert.Panics(t, func() {
		NewMessageSigner("mnemonic", config, mintControllerMap, ethNetworks)
	})
}

func TestNewMessageSigner_EthClientError(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	mnemonic := "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	config := models.CosmosNetworkConfig{
		StartBlockHeight:   1,
		Confirmations:      1,
		RPCURL:             "http://localhost:36657",
		GRPCEnabled:        true,
		GRPCHost:           "localhost",
		GRPCPort:           9090,
		TimeoutMS:          1000,
		ChainID:            "poktroll",
		ChainName:          "Poktroll",
		TxFee:              1000,
		Bech32Prefix:       "pokt",
		CoinDenom:          "upokt",
		MultisigAddress:    "pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar",
		MultisigPublicKeys: []string{"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9", "02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2", "02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6df"},
		MultisigThreshold:  2,
		MessageMonitor: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageSigner: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageRelayer: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
	}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = []byte("mintControllerAddress")

	ethNetworks := []models.EthereumNetworkConfig{
		{
			StartBlockHeight:      1,
			Confirmations:         1,
			RPCURL:                "http://localhost:8545",
			TimeoutMS:             1000,
			ChainID:               1,
			ChainName:             "Ethereum",
			MailboxAddress:        "0x0000000000000000000000000000000000000000",
			MintControllerAddress: "0x0000000000000000000000000000000000000000",
			OmniTokenAddress:      "0x0000000000000000000000000000000000000000",
			WarpISMAddress:        "0x0000000000000000000000000000000000000000",
			OracleAddresses:       []string{"0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb", "0x958d1F55E14Cba24a077b9634F16f83565fc9411", "0x4c672Edd2ec8eac8f0F1709f33de9A2E786e6912"},
			MessageMonitor: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageSigner: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageRelayer: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
		},
	}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := dbMocks.NewMockDB(t)

	// Mocking client methods
	// mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	originalNewDB := dbNewDB
	defer func() { dbNewDB = originalNewDB }()
	dbNewDB = func() db.DB {
		return mockDB
	}

	originalCosmosNewClient := cosmosNewClient
	defer func() { cosmosNewClient = originalCosmosNewClient }()
	cosmosNewClient = func(config models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
		return mockClient, nil
	}

	originEthNewClient := ethNewClient
	defer func() { ethNewClient = originEthNewClient }()
	ethNewClient = func(config models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		mockEthClient := ethMocks.NewMockEthereumClient(t)
		// mockEthClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(config.ChainID)})
		// mockEthClient.EXPECT().GetClient().Return(nil)

		return mockEthClient, assert.AnError
	}

	originalEthNewMailboxContract := ethNewMailboxContract
	defer func() { ethNewMailboxContract = originalEthNewMailboxContract }()
	ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
		return nil, nil
	}

	assert.Panics(t, func() {
		NewMessageSigner(mnemonic, config, mintControllerMap, ethNetworks)
	})
}

func TestNewMessageSigner_EthMailboxError(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	mnemonic := "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	config := models.CosmosNetworkConfig{
		StartBlockHeight:   1,
		Confirmations:      1,
		RPCURL:             "http://localhost:36657",
		GRPCEnabled:        true,
		GRPCHost:           "localhost",
		GRPCPort:           9090,
		TimeoutMS:          1000,
		ChainID:            "poktroll",
		ChainName:          "Poktroll",
		TxFee:              1000,
		Bech32Prefix:       "pokt",
		CoinDenom:          "upokt",
		MultisigAddress:    "pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar",
		MultisigPublicKeys: []string{"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9", "02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2", "02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6df"},
		MultisigThreshold:  2,
		MessageMonitor: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageSigner: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
		MessageRelayer: models.ServiceConfig{
			Enabled:    true,
			IntervalMS: 1000,
		},
	}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = []byte("mintControllerAddress")

	ethNetworks := []models.EthereumNetworkConfig{
		{
			StartBlockHeight:      1,
			Confirmations:         1,
			RPCURL:                "http://localhost:8545",
			TimeoutMS:             1000,
			ChainID:               1,
			ChainName:             "Ethereum",
			MailboxAddress:        "0x0000000000000000000000000000000000000000",
			MintControllerAddress: "0x0000000000000000000000000000000000000000",
			OmniTokenAddress:      "0x0000000000000000000000000000000000000000",
			WarpISMAddress:        "0x0000000000000000000000000000000000000000",
			OracleAddresses:       []string{"0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb", "0x958d1F55E14Cba24a077b9634F16f83565fc9411", "0x4c672Edd2ec8eac8f0F1709f33de9A2E786e6912"},
			MessageMonitor: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageSigner: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
			MessageRelayer: models.ServiceConfig{
				Enabled:    true,
				IntervalMS: 1000,
			},
		},
	}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := dbMocks.NewMockDB(t)

	// Mocking client methods
	// mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	originalNewDB := dbNewDB
	defer func() { dbNewDB = originalNewDB }()
	dbNewDB = func() db.DB {
		return mockDB
	}

	originalCosmosNewClient := cosmosNewClient
	defer func() { cosmosNewClient = originalCosmosNewClient }()
	cosmosNewClient = func(config models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
		return mockClient, nil
	}

	originEthNewClient := ethNewClient
	defer func() { ethNewClient = originEthNewClient }()
	ethNewClient = func(config models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		mockEthClient := ethMocks.NewMockEthereumClient(t)
		mockEthClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(config.ChainID)})
		mockEthClient.EXPECT().GetClient().Return(nil)

		return mockEthClient, nil
	}

	originalEthNewMailboxContract := ethNewMailboxContract
	defer func() { ethNewMailboxContract = originalEthNewMailboxContract }()
	ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
		return nil, assert.AnError
	}

	assert.Panics(t, func() {
		NewMessageSigner(mnemonic, config, mintControllerMap, ethNetworks)
	})
}
