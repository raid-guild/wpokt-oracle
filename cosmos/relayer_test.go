package cosmos

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/sirupsen/logrus"

	clientMocks "github.com/dan13ram/wpokt-oracle/cosmos/client/mocks"
	"github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/models"

	ethcommon "github.com/ethereum/go-ethereum/common"

	log "github.com/sirupsen/logrus"
)

func TestRelayerHeight(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		currentBlockHeight: 100,
	}

	height := relayer.Height()

	assert.Equal(t, uint64(100), height)
}

func TestRelayerUpdateCurrentHeight(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	relayer.UpdateCurrentHeight()

	mockClient.AssertExpectations(t)
	assert.Equal(t, uint64(100), relayer.currentBlockHeight)
}

func TestRelayerUpdateCurrentHeight_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), assert.AnError)

	relayer.UpdateCurrentHeight()

	mockClient.AssertExpectations(t)
	assert.Equal(t, uint64(0), relayer.currentBlockHeight)
}

func TestRelayerUpdateRefund(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := logrus.New().WithField("test", "relayer")

	refundID := &primitive.ObjectID{}
	update := bson.M{"status": models.RefundStatusPending}

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateRefund(refundID, update).Return(nil)

	result := relayer.UpdateRefund(refundID, update)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestRelayerUpdateRefund_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := logrus.New().WithField("test", "relayer")

	refundID := &primitive.ObjectID{}
	update := bson.M{"status": models.RefundStatusPending}

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateRefund(refundID, update).Return(assert.AnError)

	result := relayer.UpdateRefund(refundID, update)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestRelayerUpdateMessage(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := logrus.New().WithField("test", "relayer")

	messageID := &primitive.ObjectID{}
	update := bson.M{"status": models.MessageStatusPending}

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateMessage(messageID, update).Return(nil)

	result := relayer.UpdateMessage(messageID, update)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestRelayerUpdateMessage_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := logrus.New().WithField("test", "relayer")

	messageID := &primitive.ObjectID{}
	update := bson.M{"status": models.MessageStatusPending}

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateMessage(messageID, update).Return(assert.AnError)

	result := relayer.UpdateMessage(messageID, update)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateMessageTransaction(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:              &primitive.ObjectID{},
		TransactionHash: "0x010203",
		Content:         models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:      []models.Signature{},
		Sequence:        new(uint64),
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	tx := &sdk.TxResponse{}
	mockClient.EXPECT().GetTx("0x010203").Return(tx, nil)
	mockDB.EXPECT().NewCosmosTransaction(tx, mock.Anything, mock.Anything, mock.Anything, models.TransactionStatusPending).Return(models.Transaction{}, nil)
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	result := relayer.CreateMessageTransaction(message)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestCreateMessageTransaction_AddressError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	message := &models.Message{
		ID:              &primitive.ObjectID{},
		TransactionHash: "0x010203",
		Content:         models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: "recipient", Amount: "100"}},
		Signatures:      []models.Signature{},
		Sequence:        new(uint64),
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	result := relayer.CreateMessageTransaction(message)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateMessageTransaction_GetTxError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:              &primitive.ObjectID{},
		TransactionHash: "0x010203",
		Content:         models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:      []models.Signature{},
		Sequence:        new(uint64),
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	mockClient.EXPECT().GetTx("0x010203").Return(nil, assert.AnError)

	result := relayer.CreateMessageTransaction(message)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateMessageTransaction_NewError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:              &primitive.ObjectID{},
		TransactionHash: "0x010203",
		Content:         models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:      []models.Signature{},
		Sequence:        new(uint64),
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	tx := &sdk.TxResponse{}
	mockClient.EXPECT().GetTx("0x010203").Return(tx, nil)
	mockDB.EXPECT().NewCosmosTransaction(tx, mock.Anything, mock.Anything, mock.Anything, models.TransactionStatusPending).Return(models.Transaction{}, assert.AnError)

	result := relayer.CreateMessageTransaction(message)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateMessageTransaction_InsertError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:              &primitive.ObjectID{},
		TransactionHash: "0x010203",
		Content:         models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:      []models.Signature{},
		Sequence:        new(uint64),
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	tx := &sdk.TxResponse{}
	mockClient.EXPECT().GetTx("0x010203").Return(tx, nil)
	mockDB.EXPECT().NewCosmosTransaction(tx, mock.Anything, mock.Anything, mock.Anything, models.TransactionStatusPending).Return(models.Transaction{}, nil)
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, assert.AnError)

	result := relayer.CreateMessageTransaction(message)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateRefundTransaction(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:              &primitive.ObjectID{},
		TransactionHash: "txHash",
		Recipient:       recipientAddr.Hex(),
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	tx := &sdk.TxResponse{}
	mockClient.EXPECT().GetTx("txHash").Return(tx, nil)
	mockDB.EXPECT().NewCosmosTransaction(tx, mock.Anything, mock.Anything, mock.Anything, models.TransactionStatusPending).Return(models.Transaction{}, nil)
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateRefund(refund.ID, mock.Anything).Return(nil)

	result := relayer.CreateRefundTransaction(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestCreateRefundTransaction_AddressError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	refund := &models.Refund{
		ID:              &primitive.ObjectID{},
		TransactionHash: "txHash",
		Recipient:       "recipientAddr",
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	result := relayer.CreateRefundTransaction(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateRefundTransaction_GetTxError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:              &primitive.ObjectID{},
		TransactionHash: "txHash",
		Recipient:       recipientAddr.Hex(),
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	tx := &sdk.TxResponse{}
	mockClient.EXPECT().GetTx("txHash").Return(tx, assert.AnError)

	result := relayer.CreateRefundTransaction(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateRefundTransaction_NewError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:              &primitive.ObjectID{},
		TransactionHash: "txHash",
		Recipient:       recipientAddr.Hex(),
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	tx := &sdk.TxResponse{}
	mockClient.EXPECT().GetTx("txHash").Return(tx, nil)
	mockDB.EXPECT().NewCosmosTransaction(tx, mock.Anything, mock.Anything, mock.Anything, models.TransactionStatusPending).Return(models.Transaction{}, assert.AnError)

	result := relayer.CreateRefundTransaction(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateRefundTransaction_InsertError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:              &primitive.ObjectID{},
		TransactionHash: "txHash",
		Recipient:       recipientAddr.Hex(),
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	tx := &sdk.TxResponse{}
	mockClient.EXPECT().GetTx("txHash").Return(tx, nil)
	mockDB.EXPECT().NewCosmosTransaction(tx, mock.Anything, mock.Anything, mock.Anything, models.TransactionStatusPending).Return(models.Transaction{}, nil)
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, assert.AnError)

	result := relayer.CreateRefundTransaction(refund)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateTxForRefunds_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	mockDB.EXPECT().GetBroadcastedRefunds().Return(nil, assert.AnError)

	result := relayer.CreateTxForRefunds()

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateTxForRefunds(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	refund := &models.Refund{
		ID:              &primitive.ObjectID{},
		TransactionHash: "txHash",
		Recipient:       recipientAddr.Hex(),
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	mockDB.EXPECT().GetBroadcastedRefunds().Return([]models.Refund{*refund}, nil)
	tx := &sdk.TxResponse{}
	mockClient.EXPECT().GetTx("txHash").Return(tx, nil)
	mockDB.EXPECT().NewCosmosTransaction(tx, mock.Anything, mock.Anything, mock.Anything, models.TransactionStatusPending).Return(models.Transaction{}, nil)
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateRefund(refund.ID, mock.Anything).Return(nil)

	result := relayer.CreateTxForRefunds()

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestCreateTxForMessages_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	mockDB.EXPECT().GetBroadcastedMessages(mock.Anything).Return(nil, assert.AnError)

	result := relayer.CreateTxForMessages()

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateTxForMessages(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:              &primitive.ObjectID{},
		TransactionHash: "0x010203",
		Content:         models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:      []models.Signature{},
		Sequence:        new(uint64),
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	tx := &sdk.TxResponse{}
	mockDB.EXPECT().GetBroadcastedMessages(mock.Anything).Return([]models.Message{*message}, nil)
	mockClient.EXPECT().GetTx("0x010203").Return(tx, nil)
	mockDB.EXPECT().NewCosmosTransaction(tx, mock.Anything, mock.Anything, mock.Anything, models.TransactionStatusPending).Return(models.Transaction{}, nil)
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	result := relayer.CreateTxForMessages()

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestUpdateTransaction(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := logrus.New().WithField("test", "relayer")

	transaction := &models.Transaction{ID: &primitive.ObjectID{}}
	update := bson.M{"status": models.TransactionStatusConfirmed}

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateTransaction(transaction.ID, update).Return(nil)

	result := relayer.UpdateTransaction(transaction, update)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestResetRefund_Nil(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := logrus.New().WithField("test", "relayer")

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		logger: logger,
	}

	result := relayer.ResetRefund(nil)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestResetRefund(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := logrus.New().WithField("test", "relayer")

	refundID := &primitive.ObjectID{}

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		logger: logger,
	}

	update := bson.M{
		"status":           models.RefundStatusPending,
		"signatures":       []models.Signature{},
		"transaction_body": "",
		"transaction":      nil,
		"transaction_hash": "",
	}

	mockDB.EXPECT().UpdateRefund(refundID, update).Return(nil)

	result := relayer.ResetRefund(refundID)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestResetMessage_Nil(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := logrus.New().WithField("test", "relayer")

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		logger: logger,
	}

	result := relayer.ResetMessage(nil)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestResetMessage(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := logrus.New().WithField("test", "relayer")

	messageID := &primitive.ObjectID{}

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		logger: logger,
	}

	update := bson.M{
		"status":           models.MessageStatusPending,
		"signatures":       []models.Signature{},
		"transaction_body": "",
		"transaction":      nil,
		"transaction_hash": "",
	}

	mockDB.EXPECT().UpdateMessage(messageID, update).Return(nil)

	result := relayer.ResetMessage(messageID)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestConfirmTransactions_ClientError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
		multisigPk:         multisigPk,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return(nil, assert.AnError)

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, result)
}

func TestConfirmTransactions_InvalidTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	transaction := models.Transaction{
		ID:       &primitive.ObjectID{},
		Hash:     "txHash",
		Messages: []primitive.ObjectID{},
		Refund:   nil,
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
		multisigPk:         multisigPk,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{transaction}, nil)

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, result)
}

func TestConfirmTransactions_GetTxError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	transaction := models.Transaction{
		ID:       &primitive.ObjectID{},
		Hash:     "txHash",
		Messages: []primitive.ObjectID{{}},
		Refund:   nil,
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
		multisigPk:         multisigPk,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{transaction}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(nil, assert.AnError)

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, result)
}

func TestConfirmTransactions_FailedTx_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	transaction := models.Transaction{
		ID:       &primitive.ObjectID{},
		Hash:     "txHash",
		Messages: []primitive.ObjectID{{}},
		Refund:   nil,
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
		multisigPk:         multisigPk,
	}

	txResponse := &sdk.TxResponse{
		Code:   1,
		Height: 90,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{transaction}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(txResponse, nil)
	mockDB.EXPECT().UpdateTransaction(transaction.ID, bson.M{"status": models.TransactionStatusFailed}).Return(assert.AnError)

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, result)
}

func TestConfirmTransactions_FailedTx_Message(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	transaction := models.Transaction{
		ID:       &primitive.ObjectID{},
		Hash:     "txHash",
		Messages: []primitive.ObjectID{{}},
		Refund:   nil,
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
		multisigPk:         multisigPk,
	}

	txResponse := &sdk.TxResponse{
		Code:   1,
		Height: 90,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{transaction}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(txResponse, nil)
	mockDB.EXPECT().UpdateTransaction(transaction.ID, bson.M{"status": models.TransactionStatusFailed}).Return(nil)
	update := bson.M{
		"status":           models.MessageStatusPending,
		"signatures":       []models.Signature{},
		"transaction_body": "",
		"transaction":      nil,
		"transaction_hash": "",
	}
	mockDB.EXPECT().UpdateMessage(mock.Anything, update).Return(nil).Once()

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.True(t, result)
}

func TestConfirmTransactions_FailedTx_Refund(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	transaction := models.Transaction{
		ID:       &primitive.ObjectID{},
		Hash:     "txHash",
		Messages: []primitive.ObjectID{},
		Refund:   &primitive.ObjectID{},
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
		multisigPk:         multisigPk,
	}

	txResponse := &sdk.TxResponse{
		Code:   1,
		Height: 90,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{transaction}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(txResponse, nil)
	mockDB.EXPECT().UpdateTransaction(transaction.ID, bson.M{"status": models.TransactionStatusFailed}).Return(nil)
	update := bson.M{
		"status":           models.RefundStatusPending,
		"signatures":       []models.Signature{},
		"transaction_body": "",
		"transaction":      nil,
		"transaction_hash": "",
	}
	mockDB.EXPECT().UpdateRefund(mock.Anything, update).Return(nil)

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.True(t, result)
}

func TestConfirmTransactions_NotConfirmed(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	transaction := models.Transaction{
		ID:       &primitive.ObjectID{},
		Hash:     "txHash",
		Messages: []primitive.ObjectID{{}},
		Refund:   nil,
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
		multisigPk:         multisigPk,
		config:             models.CosmosNetworkConfig{Confirmations: 10},
	}

	txResponse := &sdk.TxResponse{
		Code:   0,
		Height: 100,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{transaction}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(txResponse, nil)
	update := bson.M{
		"status":        models.TransactionStatusPending,
		"confirmations": uint64(0),
	}
	mockDB.EXPECT().UpdateTransaction(transaction.ID, update).Return(nil)

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.True(t, result)
}

func TestConfirmTransactions_NotConfirmed_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	transaction := models.Transaction{
		ID:       &primitive.ObjectID{},
		Hash:     "txHash",
		Messages: []primitive.ObjectID{{}},
		Refund:   nil,
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
		multisigPk:         multisigPk,
		config:             models.CosmosNetworkConfig{Confirmations: 10},
	}

	txResponse := &sdk.TxResponse{
		Code:   0,
		Height: 100,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{transaction}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(txResponse, nil)
	update := bson.M{
		"status":        models.TransactionStatusPending,
		"confirmations": uint64(0),
	}
	mockDB.EXPECT().UpdateTransaction(transaction.ID, update).Return(assert.AnError)

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, result)
}

func TestConfirmTransactions_Confirmed_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	transaction := models.Transaction{
		ID:       &primitive.ObjectID{},
		Hash:     "txHash",
		Messages: []primitive.ObjectID{{}},
		Refund:   nil,
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
		multisigPk:         multisigPk,
		config:             models.CosmosNetworkConfig{Confirmations: 10},
	}

	txResponse := &sdk.TxResponse{
		Code:   0,
		Height: 90,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{transaction}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(txResponse, nil)
	update := bson.M{
		"status":        models.TransactionStatusConfirmed,
		"confirmations": uint64(10),
	}
	mockDB.EXPECT().UpdateTransaction(transaction.ID, update).Return(assert.AnError)

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, result)
}

func TestConfirmTransactions_Message(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	transaction := models.Transaction{
		ID:       &primitive.ObjectID{},
		Hash:     "txHash",
		Messages: []primitive.ObjectID{{}, {}},
		Refund:   nil,
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
		multisigPk:         multisigPk,
		config:             models.CosmosNetworkConfig{Confirmations: 10},
	}

	txResponse := &sdk.TxResponse{
		Code:   0,
		Height: 90,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{transaction}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(txResponse, nil)
	txUpdate := bson.M{
		"status":        models.TransactionStatusConfirmed,
		"confirmations": uint64(10),
	}
	mockDB.EXPECT().UpdateTransaction(transaction.ID, txUpdate).Return(nil)
	msgUpdate := bson.M{
		"status":           models.MessageStatusSuccess,
		"transaction":      &primitive.ObjectID{},
		"transaction_hash": "txHash",
	}
	mockDB.EXPECT().UpdateMessage(mock.Anything, msgUpdate).Return(nil).Twice()

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.True(t, result)
}

func TestConfirmTransactions_Refund(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	transaction := models.Transaction{
		ID:       &primitive.ObjectID{},
		Hash:     "txHash",
		Messages: []primitive.ObjectID{},
		Refund:   &primitive.ObjectID{},
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
		multisigPk:         multisigPk,
		config:             models.CosmosNetworkConfig{Confirmations: 10},
	}

	txResponse := &sdk.TxResponse{
		Code:   0,
		Height: 90,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{transaction}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(txResponse, nil)
	txUpdate := bson.M{
		"status":        models.TransactionStatusConfirmed,
		"confirmations": uint64(10),
	}
	mockDB.EXPECT().UpdateTransaction(transaction.ID, txUpdate).Return(nil)
	refundUpdate := bson.M{
		"status":           models.MessageStatusSuccess,
		"transaction":      &primitive.ObjectID{},
		"transaction_hash": "txHash",
	}
	mockDB.EXPECT().UpdateRefund(mock.Anything, refundUpdate).Return(nil)

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.True(t, result)
}

func TestConfirmTransactions_Refund_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	transaction := models.Transaction{
		ID:       &primitive.ObjectID{},
		Hash:     "txHash",
		Messages: []primitive.ObjectID{},
		Refund:   &primitive.ObjectID{},
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
		multisigPk:         multisigPk,
		config:             models.CosmosNetworkConfig{Confirmations: 10},
	}

	txResponse := &sdk.TxResponse{
		Code:   0,
		Height: 90,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{transaction}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(txResponse, nil)
	txUpdate := bson.M{
		"status":        models.TransactionStatusConfirmed,
		"confirmations": uint64(10),
	}
	mockDB.EXPECT().UpdateTransaction(transaction.ID, txUpdate).Return(nil)
	refundUpdate := bson.M{
		"status":           models.MessageStatusSuccess,
		"transaction":      &primitive.ObjectID{},
		"transaction_hash": "txHash",
	}
	mockDB.EXPECT().UpdateRefund(mock.Anything, refundUpdate).Return(assert.AnError)

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, result)
}

func TestRelayerInitStartBlockHeight(t *testing.T) {
	logger := logrus.New().WithField("test", "relayer")
	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

	relayer := &CosmosMessageRelayerRunnable{
		logger:             logger,
		startBlockHeight:   0,
		currentBlockHeight: 200,
	}

	relayer.InitStartBlockHeight(lastHealth)

	assert.Equal(t, uint64(100), relayer.startBlockHeight)

	relayer = &CosmosMessageRelayerRunnable{
		logger:             logger,
		startBlockHeight:   0,
		currentBlockHeight: 200,
	}
	lastHealth = nil
	relayer.InitStartBlockHeight(lastHealth)
	assert.Equal(t, uint64(200), relayer.startBlockHeight)

	relayer = &CosmosMessageRelayerRunnable{
		logger:             logger,
		startBlockHeight:   0,
		currentBlockHeight: 200,
	}
	lastHealth = &models.RunnerServiceStatus{BlockHeight: 300}
	relayer.InitStartBlockHeight(lastHealth)
	assert.Equal(t, uint64(200), relayer.startBlockHeight)
}

func TestRelayerRun(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")
	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	relayer := &CosmosMessageRelayerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)
	mockDB.EXPECT().GetBroadcastedRefunds().Return([]models.Refund{}, nil)
	mockDB.EXPECT().GetBroadcastedMessages(mock.Anything).Return([]models.Message{}, nil)
	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{}, nil)

	relayer.Run()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestNewMessageRelayer(t *testing.T) {
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

	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := mocks.NewMockDB(t)

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

	runnable := NewMessageRelayer(config, lastHealth)

	assert.NotNil(t, runnable)
	relayer, ok := runnable.(*CosmosMessageRelayerRunnable)
	assert.True(t, ok)

	assert.Equal(t, uint64(100), relayer.startBlockHeight)
	assert.Equal(t, uint64(100), relayer.currentBlockHeight)
	assert.Equal(t, config, relayer.config)
	assert.Equal(t, util.ParseChain(config), relayer.chain)
	assert.NotNil(t, relayer.client)
	assert.NotNil(t, relayer.logger)
	assert.NotNil(t, relayer.db)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestNewMessageRelayer_Disabled(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

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
			Enabled:    false,
			IntervalMS: 1000,
		},
	}

	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := mocks.NewMockDB(t)

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

	assert.Panics(t, func() {
		NewMessageRelayer(config, lastHealth)
	})

}

func TestNewMessageRelayer_InvalidPublicKey(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

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
		MultisigPublicKeys: []string{"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9", "02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2", "02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143"},
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

	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := mocks.NewMockDB(t)

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

	assert.Panics(t, func() {
		NewMessageRelayer(config, lastHealth)
	})

}

func TestNewMessageRelayer_InvalidMultisigAddress(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

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
		MultisigAddress:    "pokt1",
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

	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := mocks.NewMockDB(t)

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

	assert.Panics(t, func() {
		NewMessageRelayer(config, lastHealth)
	})

}

func TestNewMessageRelayer_ClientError(t *testing.T) {
	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

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

	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

	mockClient := clientMocks.NewMockCosmosClient(t)
	mockDB := mocks.NewMockDB(t)

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

	assert.Panics(t, func() {
		NewMessageRelayer(config, lastHealth)
	})

}
