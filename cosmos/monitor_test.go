package cosmos

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	clientMocks "github.com/dan13ram/wpokt-oracle/cosmos/client/mocks"
	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	dbMocks "github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/models"

	ethcommon "github.com/ethereum/go-ethereum/common"

	log "github.com/sirupsen/logrus"
)

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

	success := monitor.SyncNewTxs()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, success)
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

	mockDB.EXPECT().UpdateTransaction(&primitive.ObjectID{}, bson.M{
		"confirmations": uint64(2),
		"status":        models.TransactionStatusConfirmed,
	}).Return(nil)

	valid := monitor.ValidateAndConfirmTx(txDoc)

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	assert.True(t, valid)
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

	success := monitor.ConfirmTxs()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.True(t, success)
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
