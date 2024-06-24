package cosmos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"

	clientMocks "github.com/dan13ram/wpokt-oracle/cosmos/client/mocks"
	"github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/models"

	ethcommon "github.com/ethereum/go-ethereum/common"
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

func TestUpdateRefund(t *testing.T) {
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

/*
func TestCreateTxForRefunds(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	refund := &models.Refund{
		ID:              &primitive.ObjectID{},
		TransactionHash: "txHash",
		Recipient:       "recipient",
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockDB.EXPECT().GetBroadcastedRefunds().Return([]models.Refund{*refund}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(&sdk.TxResponse{}, nil)
	mockDB.EXPECT().NewCosmosTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, models.TransactionStatusPending).Return(models.Transaction{}, nil)
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateRefund(refund.ID, mock.Anything).Return(nil)

	result := relayer.CreateTxForRefunds()

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestCreateTxForMessages(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	message := &models.Message{
		ID:              &primitive.ObjectID{},
		TransactionHash: "txHash",
		Content:         models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockDB.EXPECT().GetBroadcastedMessages(mock.Anything).Return([]models.Message{*message}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(&sdk.TxResponse{}, nil)
	mockDB.EXPECT().NewCosmosTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, models.TransactionStatusPending).Return(models.Transaction{}, nil)
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

func TestConfirmTransactions(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "relayer")

	transaction := models.Transaction{
		ID:       &primitive.ObjectID{},
		Hash:     "txHash",
		Messages: []primitive.ObjectID{primitive.ObjectID{}},
		Refund:   nil,
	}

	relayer := &CosmosMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
	}

	txResponse := &sdk.TxResponse{
		Code:   0,
		Height: 90,
	}

	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{transaction}, nil)
	mockClient.EXPECT().GetTx("txHash").Return(txResponse, nil)
	mockDB.EXPECT().UpdateTransaction(transaction.ID, mock.Anything).Return(nil)
	mockDB.EXPECT().UpdateMessage(mock.Anything, mock.Anything).Return(nil)

	result := relayer.ConfirmTransactions()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.True(t, result)
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

	relayer := &CosmosMessageRelayerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)
	mockDB.EXPECT().GetBroadcastedRefunds().Return([]models.Refund{}, nil)
	mockDB.EXPECT().GetBroadcastedMessages(mock.Anything).Return([]models.Message{}, nil)
	mockDB.EXPECT().GetPendingTransactionsFrom(mock.Anything, mock.Anything).Return([]models.Transaction{}, nil)

	relayer.Run()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}
*/
