package cosmos

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	clientMocks "github.com/dan13ram/wpokt-oracle/cosmos/client/mocks"
	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/db"
	dbMocks "github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/models"

	ethcommon "github.com/ethereum/go-ethereum/common"

	log "github.com/sirupsen/logrus"
)

func TestMonitorHeight(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	monitor := &CosmosMessageMonitorRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
	}

	height := monitor.Height()

	assert.Equal(t, uint64(100), height)
}

func TestMonitorUpdateCurrentHeight(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	monitor.UpdateCurrentHeight()

	mockClient.AssertExpectations(t)
	assert.Equal(t, uint64(100), monitor.currentBlockHeight)
}

func TestMonitorUpdateCurrentHeight_Error(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), assert.AnError)

	monitor.UpdateCurrentHeight()

	mockClient.AssertExpectations(t)
	assert.Equal(t, uint64(0), monitor.currentBlockHeight)
}

func TestCreateRefund(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

	txRes := &sdk.TxResponse{}
	txDoc := &models.Transaction{}
	toAddr := []byte("some-address")
	amount := sdk.NewCoin("token", math.NewInt(100))

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().NewRefund(txRes, txDoc, toAddr, amount).Return(models.Refund{}, nil)
	mockDB.EXPECT().InsertRefund(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateTransaction(txDoc.ID, mock.Anything).Return(nil)

	result := monitor.CreateRefund(txRes, txDoc, toAddr, amount)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestCreateRefund_NewError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

	txRes := &sdk.TxResponse{}
	txDoc := &models.Transaction{}
	toAddr := []byte("some-address")
	amount := sdk.NewCoin("token", math.NewInt(100))

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().NewRefund(txRes, txDoc, toAddr, amount).Return(models.Refund{}, assert.AnError)

	result := monitor.CreateRefund(txRes, txDoc, toAddr, amount)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateRefund_InsertError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

	txRes := &sdk.TxResponse{}
	txDoc := &models.Transaction{}
	toAddr := []byte("some-address")
	amount := sdk.NewCoin("token", math.NewInt(100))

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().NewRefund(txRes, txDoc, toAddr, amount).Return(models.Refund{}, nil)
	mockDB.EXPECT().InsertRefund(mock.Anything).Return(primitive.ObjectID{}, assert.AnError)

	result := monitor.CreateRefund(txRes, txDoc, toAddr, amount)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateRefund_UpdateError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

	txRes := &sdk.TxResponse{}
	txDoc := &models.Transaction{}
	toAddr := []byte("some-address")
	amount := sdk.NewCoin("token", math.NewInt(100))

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().NewRefund(txRes, txDoc, toAddr, amount).Return(models.Refund{}, nil)
	mockDB.EXPECT().InsertRefund(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateTransaction(txDoc.ID, mock.Anything).Return(assert.AnError)

	result := monitor.CreateRefund(txRes, txDoc, toAddr, amount)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateMessage(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")
	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))
	amountCoin := sdk.NewCoin("token", math.NewInt(100))
	memo := models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"}

	txRes := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}
	txDoc := &models.Transaction{}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	mockDB.EXPECT().NewMessageBody(senderAddress[:], amountCoin.Amount.BigInt(), recipientAddress[:]).Return(models.MessageBody{}, nil)
	mockDB.EXPECT().NewMessageContent(uint32(1), uint32(0), senderAddress[:], uint32(1), mintControllerAddress[:], models.MessageBody{}).Return(models.MessageContent{}, nil)
	mockDB.EXPECT().NewMessage(txDoc, mock.Anything, models.MessageStatusPending).Return(models.Message{}, nil)
	mockDB.EXPECT().InsertMessage(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateTransaction(txDoc.ID, mock.Anything).Return(nil)

	result := monitor.CreateMessage(txRes, tx, txDoc, senderAddress[:], amountCoin, memo)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestCreateMessage_AddressError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")
	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))
	amountCoin := sdk.NewCoin("token", math.NewInt(100))
	memo := models.MintMemo{Address: "0xaddress", ChainID: "1"}

	txRes := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}
	txDoc := &models.Transaction{}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	result := monitor.CreateMessage(txRes, tx, txDoc, senderAddress[:], amountCoin, memo)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateMessage_NewBodyError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")
	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))
	amountCoin := sdk.NewCoin("token", math.NewInt(100))
	memo := models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"}

	txRes := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}
	txDoc := &models.Transaction{}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	mockDB.EXPECT().NewMessageBody(senderAddress[:], amountCoin.Amount.BigInt(), recipientAddress[:]).Return(models.MessageBody{}, assert.AnError)

	result := monitor.CreateMessage(txRes, tx, txDoc, senderAddress[:], amountCoin, memo)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateMessage_SignerInfoError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")
	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))
	amountCoin := sdk.NewCoin("token", math.NewInt(100))
	memo := models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"}

	txRes := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{}}}
	txDoc := &models.Transaction{}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	mockDB.EXPECT().NewMessageBody(senderAddress[:], amountCoin.Amount.BigInt(), recipientAddress[:]).Return(models.MessageBody{}, nil)

	result := monitor.CreateMessage(txRes, tx, txDoc, senderAddress[:], amountCoin, memo)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateMessage_NewContentError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")
	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))
	amountCoin := sdk.NewCoin("token", math.NewInt(100))
	memo := models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"}

	txRes := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}
	txDoc := &models.Transaction{}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	mockDB.EXPECT().NewMessageBody(senderAddress[:], amountCoin.Amount.BigInt(), recipientAddress[:]).Return(models.MessageBody{}, nil)
	mockDB.EXPECT().NewMessageContent(uint32(1), uint32(0), senderAddress[:], uint32(1), mintControllerAddress[:], models.MessageBody{}).Return(models.MessageContent{}, assert.AnError)

	result := monitor.CreateMessage(txRes, tx, txDoc, senderAddress[:], amountCoin, memo)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateMessage_MintControllerError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")
	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	amountCoin := sdk.NewCoin("token", math.NewInt(100))
	memo := models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"}

	txRes := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}
	txDoc := &models.Transaction{}

	mintControllerMap := make(map[uint32][]byte)

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	mockDB.EXPECT().NewMessageBody(senderAddress[:], amountCoin.Amount.BigInt(), recipientAddress[:]).Return(models.MessageBody{}, nil)

	result := monitor.CreateMessage(txRes, tx, txDoc, senderAddress[:], amountCoin, memo)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateMessage_NewMessageError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")
	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))
	amountCoin := sdk.NewCoin("token", math.NewInt(100))
	memo := models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"}

	txRes := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}
	txDoc := &models.Transaction{}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	mockDB.EXPECT().NewMessageBody(senderAddress[:], amountCoin.Amount.BigInt(), recipientAddress[:]).Return(models.MessageBody{}, nil)
	mockDB.EXPECT().NewMessageContent(uint32(1), uint32(0), senderAddress[:], uint32(1), mintControllerAddress[:], models.MessageBody{}).Return(models.MessageContent{}, nil)
	mockDB.EXPECT().NewMessage(txDoc, mock.Anything, models.MessageStatusPending).Return(models.Message{}, assert.AnError)

	result := monitor.CreateMessage(txRes, tx, txDoc, senderAddress[:], amountCoin, memo)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateMessage_InsertError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")
	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))
	amountCoin := sdk.NewCoin("token", math.NewInt(100))
	memo := models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"}

	txRes := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}
	txDoc := &models.Transaction{}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	mockDB.EXPECT().NewMessageBody(senderAddress[:], amountCoin.Amount.BigInt(), recipientAddress[:]).Return(models.MessageBody{}, nil)
	mockDB.EXPECT().NewMessageContent(uint32(1), uint32(0), senderAddress[:], uint32(1), mintControllerAddress[:], models.MessageBody{}).Return(models.MessageContent{}, nil)
	mockDB.EXPECT().NewMessage(txDoc, mock.Anything, models.MessageStatusPending).Return(models.Message{}, nil)
	mockDB.EXPECT().InsertMessage(mock.Anything).Return(primitive.ObjectID{}, assert.AnError)

	result := monitor.CreateMessage(txRes, tx, txDoc, senderAddress[:], amountCoin, memo)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestCreateMessage_UpdateTxError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")
	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))
	amountCoin := sdk.NewCoin("token", math.NewInt(100))
	memo := models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"}

	txRes := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}
	txDoc := &models.Transaction{}

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	mockDB.EXPECT().NewMessageBody(senderAddress[:], amountCoin.Amount.BigInt(), recipientAddress[:]).Return(models.MessageBody{}, nil)
	mockDB.EXPECT().NewMessageContent(uint32(1), uint32(0), senderAddress[:], uint32(1), mintControllerAddress[:], models.MessageBody{}).Return(models.MessageContent{}, nil)
	mockDB.EXPECT().NewMessage(txDoc, mock.Anything, models.MessageStatusPending).Return(models.Message{}, nil)
	mockDB.EXPECT().InsertMessage(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateTransaction(txDoc.ID, mock.Anything).Return(assert.AnError)

	result := monitor.CreateMessage(txRes, tx, txDoc, senderAddress[:], amountCoin, memo)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSyncNewTxs(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")
	multisigAddress := ethcommon.BytesToAddress([]byte("multisigAddress"))

	monitor := &CosmosMessageMonitorRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		startBlockHeight:   1,
		currentBlockHeight: 10,
		config: models.CosmosNetworkConfig{
			MultisigAddress: multisigAddress.Hex(),
		},
		multisigAddressBytes: multisigAddress.Bytes(),
	}

	txResponses := []*sdk.TxResponse{
		{TxHash: "tx1"},
		{TxHash: "tx2"},
	}

	mockClient.EXPECT().GetTxsSentToAddressAfterHeight(multisigAddress.Hex(), uint64(1)).Return(txResponses, nil).Once()
	mockDB.EXPECT().NewCosmosTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.Transaction{}, nil).Twice()
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, nil).Twice()
	result := &util.ValidateTxResult{
		Confirmations: 0,
		TxStatus:      models.TransactionStatusPending,
	}

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	success := monitor.SyncNewTxs()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, success)
}

func TestSyncNewTxs_NoNewBlocks(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")
	multisigAddress := ethcommon.BytesToAddress([]byte("multisigAddress"))

	monitor := &CosmosMessageMonitorRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		startBlockHeight:   10,
		currentBlockHeight: 10,
		config: models.CosmosNetworkConfig{
			MultisigAddress: multisigAddress.Hex(),
		},
		multisigAddressBytes: multisigAddress.Bytes(),
	}

	success := monitor.SyncNewTxs()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, success)
}

func TestSyncNewTxs_ClientError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")
	multisigAddress := ethcommon.BytesToAddress([]byte("multisigAddress"))

	monitor := &CosmosMessageMonitorRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		startBlockHeight:   1,
		currentBlockHeight: 10,
		config: models.CosmosNetworkConfig{
			MultisigAddress: multisigAddress.Hex(),
		},
		multisigAddressBytes: multisigAddress.Bytes(),
	}

	txResponses := []*sdk.TxResponse{
		{TxHash: "tx1"},
		{TxHash: "tx2"},
	}

	mockClient.EXPECT().GetTxsSentToAddressAfterHeight(multisigAddress.Hex(), uint64(1)).Return(txResponses, assert.AnError).Once()

	success := monitor.SyncNewTxs()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, success)
}

func TestSyncNewTxs_ValidateError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")
	multisigAddress := ethcommon.BytesToAddress([]byte("multisigAddress"))

	monitor := &CosmosMessageMonitorRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		startBlockHeight:   1,
		currentBlockHeight: 10,
		config: models.CosmosNetworkConfig{
			MultisigAddress: multisigAddress.Hex(),
		},
		multisigAddressBytes: multisigAddress.Bytes(),
	}

	txResponses := []*sdk.TxResponse{
		{TxHash: "tx1"},
		{TxHash: "tx2"},
	}

	mockClient.EXPECT().GetTxsSentToAddressAfterHeight(multisigAddress.Hex(), uint64(1)).Return(txResponses, nil).Once()
	result := &util.ValidateTxResult{
		Confirmations: 0,
		TxStatus:      models.TransactionStatusPending,
	}

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, assert.AnError
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	success := monitor.SyncNewTxs()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, success)
}

func TestSyncNewTxs_NewError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")
	multisigAddress := ethcommon.BytesToAddress([]byte("multisigAddress"))

	monitor := &CosmosMessageMonitorRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		startBlockHeight:   1,
		currentBlockHeight: 10,
		config: models.CosmosNetworkConfig{
			MultisigAddress: multisigAddress.Hex(),
		},
		multisigAddressBytes: multisigAddress.Bytes(),
	}

	txResponses := []*sdk.TxResponse{
		{TxHash: "tx1"},
		{TxHash: "tx2"},
	}

	mockClient.EXPECT().GetTxsSentToAddressAfterHeight(multisigAddress.Hex(), uint64(1)).Return(txResponses, nil).Once()
	mockDB.EXPECT().NewCosmosTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.Transaction{}, assert.AnError).Once()
	mockDB.EXPECT().NewCosmosTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.Transaction{}, nil).Once()
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, nil).Once()
	result := &util.ValidateTxResult{
		Confirmations: 0,
		TxStatus:      models.TransactionStatusPending,
	}

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	success := monitor.SyncNewTxs()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, success)
}

func TestSyncNewTxs_InsertError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")
	multisigAddress := ethcommon.BytesToAddress([]byte("multisigAddress"))

	monitor := &CosmosMessageMonitorRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		startBlockHeight:   1,
		currentBlockHeight: 10,
		config: models.CosmosNetworkConfig{
			MultisigAddress: multisigAddress.Hex(),
		},
		multisigAddressBytes: multisigAddress.Bytes(),
	}

	txResponses := []*sdk.TxResponse{
		{TxHash: "tx1"},
		{TxHash: "tx2"},
	}

	mockClient.EXPECT().GetTxsSentToAddressAfterHeight(multisigAddress.Hex(), uint64(1)).Return(txResponses, nil).Once()
	mockDB.EXPECT().NewCosmosTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.Transaction{}, nil).Twice()
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, assert.AnError).Once()
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, nil).Once()
	result := &util.ValidateTxResult{
		Confirmations: 0,
		TxStatus:      models.TransactionStatusPending,
	}

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	success := monitor.SyncNewTxs()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, success)
}

func TestValidateAndConfirmTx(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "hash1",
	}

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	txResponse := &sdk.TxResponse{}

	mockClient.EXPECT().GetTx("hash1").Return(txResponse, nil)

	result := &util.ValidateTxResult{
		Confirmations: 2,
		TxStatus:      models.TransactionStatusConfirmed,
	}

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().UpdateTransaction(&primitive.ObjectID{}, bson.M{
		"confirmations": uint64(2),
		"status":        models.TransactionStatusConfirmed,
	}).Return(nil)

	valid := monitor.ValidateAndConfirmTx(txDoc)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, valid)
}

func TestValidateAndConfirmTx_ClientError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "hash1",
	}

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	txResponse := &sdk.TxResponse{}

	mockClient.EXPECT().GetTx("hash1").Return(txResponse, assert.AnError)

	valid := monitor.ValidateAndConfirmTx(txDoc)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, valid)
}

func TestValidateAndConfirmTx_ValidateError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "hash1",
	}

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	txResponse := &sdk.TxResponse{}

	mockClient.EXPECT().GetTx("hash1").Return(txResponse, nil)

	result := &util.ValidateTxResult{
		Confirmations: 2,
		TxStatus:      models.TransactionStatusConfirmed,
	}

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, assert.AnError
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	valid := monitor.ValidateAndConfirmTx(txDoc)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, valid)
}

func TestValidateAndConfirmTx_UpdateError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "hash1",
	}

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	txResponse := &sdk.TxResponse{}

	mockClient.EXPECT().GetTx("hash1").Return(txResponse, nil)

	result := &util.ValidateTxResult{
		Confirmations: 2,
		TxStatus:      models.TransactionStatusConfirmed,
	}

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().UpdateTransaction(&primitive.ObjectID{}, bson.M{
		"confirmations": uint64(2),
		"status":        models.TransactionStatusConfirmed,
	}).Return(assert.AnError)

	valid := monitor.ValidateAndConfirmTx(txDoc)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.False(t, valid)
}

func TestConfirmTxs(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	txs := []models.Transaction{
		{ID: &primitive.ObjectID{}, Hash: "hash1"},
		{ID: &primitive.ObjectID{}, Hash: "hash2"},
	}

	mockDB.EXPECT().GetPendingTransactionsTo(mock.Anything, mock.Anything).Return(txs, nil)
	mockClient.EXPECT().GetTx("hash1").Return(&sdk.TxResponse{}, nil)
	mockClient.EXPECT().GetTx("hash2").Return(&sdk.TxResponse{}, nil)

	mockDB.EXPECT().UpdateTransaction(&primitive.ObjectID{}, mock.Anything).Return(nil)
	mockDB.EXPECT().UpdateTransaction(&primitive.ObjectID{}, mock.Anything).Return(nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return &util.ValidateTxResult{
			Confirmations: 2,
			TxStatus:      models.TransactionStatusConfirmed,
		}, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	success := monitor.ConfirmTxs()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.True(t, success)
}

func TestConfirmTxs_ClientError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	txs := []models.Transaction{
		{ID: &primitive.ObjectID{}, Hash: "hash1"},
		{ID: &primitive.ObjectID{}, Hash: "hash2"},
	}

	mockDB.EXPECT().GetPendingTransactionsTo(mock.Anything, mock.Anything).Return(txs, assert.AnError)

	success := monitor.ConfirmTxs()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, success)
}

func TestValidateTxAndCreate(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))
	amountCoin := sdk.NewCoin("token", math.NewInt(100))

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "hash1",
	}

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	txResponse := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}

	result := &util.ValidateTxResult{
		Confirmations: 2,
		TxStatus:      models.TransactionStatusConfirmed,
		NeedsRefund:   false,
		SenderAddress: senderAddress.Bytes(),
		Amount:        sdk.NewCoin("token", math.NewInt(100)),
		Memo:          models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"},
		Tx:            tx,
	}

	mockClient.EXPECT().GetTx("hash1").Return(txResponse, nil)
	mockDB.EXPECT().LockWriteTransaction(txDoc).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().NewMessageBody(senderAddress[:], amountCoin.Amount.BigInt(), recipientAddress[:]).Return(models.MessageBody{}, nil)
	mockDB.EXPECT().NewMessageContent(uint32(1), uint32(0), senderAddress[:], uint32(1), mintControllerAddress[:], models.MessageBody{}).Return(models.MessageContent{}, nil)
	mockDB.EXPECT().NewMessage(txDoc, mock.Anything, models.MessageStatusPending).Return(models.Message{}, nil)
	mockDB.EXPECT().InsertMessage(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateTransaction(txDoc.ID, mock.Anything).Return(nil)

	success := monitor.ValidateTxAndCreate(txDoc)

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.True(t, success)
}

func TestValidateTxAndCreate_ClientError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "hash1",
	}

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	txResponse := &sdk.TxResponse{}

	mockClient.EXPECT().GetTx("hash1").Return(txResponse, assert.AnError)

	success := monitor.ValidateTxAndCreate(txDoc)

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, success)
}

func TestValidateTxAndCreate_ValidateError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "hash1",
	}

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	txResponse := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}

	result := &util.ValidateTxResult{
		Confirmations: 2,
		TxStatus:      models.TransactionStatusConfirmed,
		NeedsRefund:   false,
		SenderAddress: senderAddress.Bytes(),
		Amount:        sdk.NewCoin("token", math.NewInt(100)),
		Memo:          models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"},
		Tx:            tx,
	}

	mockClient.EXPECT().GetTx("hash1").Return(txResponse, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, assert.AnError
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	success := monitor.ValidateTxAndCreate(txDoc)

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, success)
}

func TestValidateTxAndCreate_TxPending(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "hash1",
	}

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	txResponse := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}

	result := &util.ValidateTxResult{
		Confirmations: 2,
		TxStatus:      models.TransactionStatusPending,
		NeedsRefund:   false,
		SenderAddress: senderAddress.Bytes(),
		Amount:        sdk.NewCoin("token", math.NewInt(100)),
		Memo:          models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"},
		Tx:            tx,
	}

	mockClient.EXPECT().GetTx("hash1").Return(txResponse, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	success := monitor.ValidateTxAndCreate(txDoc)

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, success)
}

func TestValidateTxAndCreate_TxInvalid(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "hash1",
	}

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	txResponse := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}

	result := &util.ValidateTxResult{
		Confirmations: 2,
		TxStatus:      models.TransactionStatusInvalid,
		NeedsRefund:   false,
		SenderAddress: senderAddress.Bytes(),
		Amount:        sdk.NewCoin("token", math.NewInt(100)),
		Memo:          models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"},
		Tx:            tx,
	}

	mockClient.EXPECT().GetTx("hash1").Return(txResponse, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().UpdateTransaction(txDoc.ID, bson.M{"status": result.TxStatus}).Return(nil)

	success := monitor.ValidateTxAndCreate(txDoc)

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.True(t, success)
}

func TestValidateTxAndCreate_TxInvalidError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "hash1",
	}

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	txResponse := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}

	result := &util.ValidateTxResult{
		Confirmations: 2,
		TxStatus:      models.TransactionStatusInvalid,
		NeedsRefund:   false,
		SenderAddress: senderAddress.Bytes(),
		Amount:        sdk.NewCoin("token", math.NewInt(100)),
		Memo:          models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"},
		Tx:            tx,
	}

	mockClient.EXPECT().GetTx("hash1").Return(txResponse, nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().UpdateTransaction(txDoc.ID, bson.M{"status": result.TxStatus}).Return(assert.AnError)

	success := monitor.ValidateTxAndCreate(txDoc)

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, success)
}

func TestValidateTxAndCreate_LockError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "hash1",
	}

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	txResponse := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}

	result := &util.ValidateTxResult{
		Confirmations: 2,
		TxStatus:      models.TransactionStatusConfirmed,
		NeedsRefund:   false,
		SenderAddress: senderAddress.Bytes(),
		Amount:        sdk.NewCoin("token", math.NewInt(100)),
		Memo:          models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"},
		Tx:            tx,
	}

	mockClient.EXPECT().GetTx("hash1").Return(txResponse, nil)
	mockDB.EXPECT().LockWriteTransaction(txDoc).Return("lock-id", assert.AnError)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	success := monitor.ValidateTxAndCreate(txDoc)

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, success)
}

func TestValidateTxAndCreate_Refund(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))

	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "hash1",
	}

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	txResponse := &sdk.TxResponse{}
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}

	result := &util.ValidateTxResult{
		Confirmations: 2,
		TxStatus:      models.TransactionStatusConfirmed,
		NeedsRefund:   true,
		SenderAddress: senderAddress.Bytes(),
		Amount:        sdk.NewCoin("token", math.NewInt(100)),
		Memo:          models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"},
		Tx:            tx,
	}

	mockClient.EXPECT().GetTx("hash1").Return(txResponse, nil)
	mockDB.EXPECT().LockWriteTransaction(txDoc).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return result, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	amount := sdk.NewCoin("token", math.NewInt(100))

	mockDB.EXPECT().NewRefund(txResponse, txDoc, senderAddress.Bytes(), amount).Return(models.Refund{}, nil)
	mockDB.EXPECT().InsertRefund(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateTransaction(txDoc.ID, mock.Anything).Return(nil)

	success := monitor.ValidateTxAndCreate(txDoc)

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.True(t, success)
}

func TestCreateRefundsOrMessagesForConfirmedTxs(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")
	senderAddress := ethcommon.BytesToAddress([]byte("sender"))
	recipientAddress := ethcommon.BytesToAddress([]byte("recipient"))
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))
	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()
	tx := &tx.Tx{AuthInfo: &tx.AuthInfo{SignerInfos: []*tx.SignerInfo{{Sequence: 1}}}}

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	txs := []models.Transaction{
		{ID: &primitive.ObjectID{}, Hash: "hash1"},
		{ID: &primitive.ObjectID{}, Hash: "hash2"},
	}

	mockDB.EXPECT().GetConfirmedTransactionsTo(mock.Anything, mock.Anything).Return(txs, nil)
	mockClient.EXPECT().GetTx("hash1").Return(&sdk.TxResponse{}, nil)
	mockClient.EXPECT().GetTx("hash2").Return(&sdk.TxResponse{}, nil)

	mockDB.EXPECT().LockWriteTransaction(mock.Anything).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
		return &util.ValidateTxResult{
			Confirmations: 2,
			TxStatus:      models.TransactionStatusConfirmed,
			NeedsRefund:   false,
			SenderAddress: senderAddress.Bytes(),
			Amount:        sdk.NewCoin("token", math.NewInt(100)),
			Memo:          models.MintMemo{Address: recipientAddress.Hex(), ChainID: "1"},
			Tx:            tx,
		}, nil
	}
	defer func() { utilValidateTxToCosmosMultisig = util.ValidateTxToCosmosMultisig }()

	mockDB.EXPECT().NewMessageBody(mock.Anything, mock.Anything, mock.Anything).Return(models.MessageBody{}, nil)
	mockDB.EXPECT().NewMessageContent(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.MessageContent{}, nil)
	mockDB.EXPECT().NewMessage(mock.Anything, mock.Anything, models.MessageStatusPending).Return(models.Message{}, nil)
	mockDB.EXPECT().InsertMessage(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateTransaction(mock.Anything, mock.Anything).Return(nil)

	success := monitor.CreateRefundsOrMessagesForConfirmedTxs()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.True(t, success)
}

func TestCreateRefundsOrMessagesForConfirmedTxs_ClientError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")
	mintControllerAddress := ethcommon.BytesToAddress([]byte("mintController"))
	mintControllerMap := make(map[uint32][]byte)
	mintControllerMap[1] = mintControllerAddress.Bytes()

	monitor := &CosmosMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
	}

	txs := []models.Transaction{
		{ID: &primitive.ObjectID{}, Hash: "hash1"},
		{ID: &primitive.ObjectID{}, Hash: "hash2"},
	}

	mockDB.EXPECT().GetConfirmedTransactionsTo(mock.Anything, mock.Anything).Return(txs, assert.AnError)
	success := monitor.CreateRefundsOrMessagesForConfirmedTxs()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.False(t, success)
}

func TestMonitorInitStartBlockHeight(t *testing.T) {
	logger := log.New().WithField("test", "monitor")
	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

	monitor := &CosmosMessageMonitorRunnable{
		logger:             logger,
		startBlockHeight:   0,
		currentBlockHeight: 200,
	}

	monitor.InitStartBlockHeight(lastHealth)

	assert.Equal(t, uint64(100), monitor.startBlockHeight)

	monitor = &CosmosMessageMonitorRunnable{
		logger:             logger,
		startBlockHeight:   0,
		currentBlockHeight: 200,
	}
	lastHealth = nil
	monitor.InitStartBlockHeight(lastHealth)
	assert.Equal(t, uint64(200), monitor.startBlockHeight)

	monitor = &CosmosMessageMonitorRunnable{
		logger:             logger,
		startBlockHeight:   0,
		currentBlockHeight: 200,
	}
	lastHealth = &models.RunnerServiceStatus{BlockHeight: 300}
	monitor.InitStartBlockHeight(lastHealth)
	assert.Equal(t, uint64(200), monitor.startBlockHeight)
}

func TestMonitorRun(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "monitor")

	monitor := &CosmosMessageMonitorRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)
	mockClient.EXPECT().GetTxsSentToAddressAfterHeight(mock.Anything, mock.Anything).Return([]*sdk.TxResponse{}, nil)
	mockDB.EXPECT().GetPendingTransactionsTo(mock.Anything, mock.Anything).Return([]models.Transaction{}, nil)
	mockDB.EXPECT().GetConfirmedTransactionsTo(mock.Anything, mock.Anything).Return([]models.Transaction{}, nil)

	monitor.Run()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestNewMessageMonitor(t *testing.T) {
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

	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

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

	runnable := NewMessageMonitor(config, mintControllerMap, ethNetworks, lastHealth)

	assert.NotNil(t, runnable)
	monitor, ok := runnable.(*CosmosMessageMonitorRunnable)
	assert.True(t, ok)

	assert.Equal(t, uint64(100), monitor.startBlockHeight)
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

func TestNewMessageMonitor_Disabled(t *testing.T) {
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
			Enabled:    false,
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

	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

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

	assert.Panics(t, func() {
		NewMessageMonitor(config, mintControllerMap, ethNetworks, lastHealth)
	})

}

func TestNewMessageMonitor_MultisigPublicKeyError(t *testing.T) {
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

	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

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

	assert.Panics(t, func() {
		NewMessageMonitor(config, mintControllerMap, ethNetworks, lastHealth)
	})

}

func TestNewMessageMonitor_MultisigAddressError(t *testing.T) {
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
		MultisigAddress:    "pokt",
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

	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

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

	assert.Panics(t, func() {
		NewMessageMonitor(config, mintControllerMap, ethNetworks, lastHealth)
	})

}

func TestNewMessageMonitor_ClientError(t *testing.T) {
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

	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

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

	assert.Panics(t, func() {
		NewMessageMonitor(config, mintControllerMap, ethNetworks, lastHealth)
	})

}
