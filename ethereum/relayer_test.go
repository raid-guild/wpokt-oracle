package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	clientMocks "github.com/dan13ram/wpokt-oracle/ethereum/client/mocks"
	"github.com/dan13ram/wpokt-oracle/models"

	log "github.com/sirupsen/logrus"
)

func TestRelayerHeight(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "relayer")

	monitor := &EthMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
	}

	height := monitor.Height()

	assert.Equal(t, uint64(100), height)
}

func TestRelayerUpdateCurrentBlockHeight_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:     mockDB,
		client: mockEthClient,
		logger: logger,
	}

	mockEthClient.EXPECT().GetBlockHeight().Return(uint64(100), assert.AnError)

	relayer.UpdateCurrentBlockHeight()

	assert.Equal(t, uint64(0), relayer.currentBlockHeight)
}

func TestRelayerUpdateCurrentBlockHeight(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:     mockDB,
		client: mockEthClient,
		logger: logger,
	}

	mockEthClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)

	relayer.UpdateCurrentBlockHeight()

	assert.Equal(t, uint64(100), relayer.currentBlockHeight)
}

func TestCreateTxForFulfillmentEvent_Nil(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	success := relayer.CreateTxForFulfillmentEvent(nil)
	assert.False(t, success)
}

func TestCreateTxForFulfillmentEvent_ValidateError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	event := &autogen.MintControllerFulfillment{
		Raw: types.Log{
			TxHash: common.HexToHash("0x1"),
		},
	}

	ethValidateTransactionByHash = func(client eth.EthereumClient, txHash string) (*ValidateTransactionByHashResult, error) {
		assert.Equal(t, mockEthClient, client)
		assert.Equal(t, common.HexToHash("0x1").Hex(), txHash)
		return &ValidateTransactionByHashResult{
			Receipt: &types.Receipt{Status: types.ReceiptStatusSuccessful},
			Tx:      &types.Transaction{},
		}, assert.AnError
	}

	success := relayer.CreateTxForFulfillmentEvent(event)
	assert.False(t, success)
}

func TestCreateTxForFulfillmentEvent_NewError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	event := &autogen.MintControllerFulfillment{
		Raw: types.Log{
			TxHash: common.HexToHash("0x1"),
		},
	}

	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))

	ethValidateTransactionByHash = func(client eth.EthereumClient, txHash string) (*ValidateTransactionByHashResult, error) {
		assert.Equal(t, mockEthClient, client)
		assert.Equal(t, common.HexToHash("0x1").Hex(), txHash)
		return &ValidateTransactionByHashResult{
			Receipt: &types.Receipt{Status: types.ReceiptStatusSuccessful},
			Tx:      &types.Transaction{},
		}, nil
	}

	mockDB.EXPECT().NewEthereumTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.Transaction{}, assert.AnError)

	success := relayer.CreateTxForFulfillmentEvent(event)
	assert.False(t, success)
}

func TestCreateTxForFulfillmentEvent_InsertError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	event := &autogen.MintControllerFulfillment{
		Raw: types.Log{
			TxHash: common.HexToHash("0x1"),
		},
	}

	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))

	ethValidateTransactionByHash = func(client eth.EthereumClient, txHash string) (*ValidateTransactionByHashResult, error) {
		assert.Equal(t, mockEthClient, client)
		assert.Equal(t, common.HexToHash("0x1").Hex(), txHash)
		return &ValidateTransactionByHashResult{
			Receipt: &types.Receipt{Status: types.ReceiptStatusSuccessful},
			Tx:      &types.Transaction{},
		}, nil
	}

	mockDB.EXPECT().NewEthereumTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.Transaction{}, nil)
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, assert.AnError)

	success := relayer.CreateTxForFulfillmentEvent(event)
	assert.False(t, success)
}

func TestCreateTxForFulfillmentEvent(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	event := &autogen.MintControllerFulfillment{
		Raw: types.Log{
			TxHash: common.HexToHash("0x1"),
		},
	}

	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))

	ethValidateTransactionByHash = func(client eth.EthereumClient, txHash string) (*ValidateTransactionByHashResult, error) {
		assert.Equal(t, mockEthClient, client)
		assert.Equal(t, common.HexToHash("0x1").Hex(), txHash)
		return &ValidateTransactionByHashResult{
			Receipt: &types.Receipt{Status: types.ReceiptStatusSuccessful},
			Tx:      &types.Transaction{},
		}, nil
	}

	mockDB.EXPECT().NewEthereumTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.Transaction{}, nil)
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, nil)

	success := relayer.CreateTxForFulfillmentEvent(event)
	assert.True(t, success)
}

func TestValidateTransactionAndParseFulfillmentEvents_ClientError(t *testing.T) {
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
	}

	txHash := "0x1"
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(90),
		Logs: []*types.Log{
			{
				Address: common.HexToAddress("0x1"),
			},
		},
	}

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, assert.AnError)

	result, err := relayer.ValidateTransactionAndParseFulfillmentEvents(txHash)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestValidateTransactionAndParseFulfillmentEvents_FailedTx(t *testing.T) {
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
	}

	txHash := "0x1"
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusFailed,
		BlockNumber: big.NewInt(90),
		Logs: []*types.Log{
			{
				Address: common.HexToAddress("0x1"),
			},
		},
	}

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil)

	result, err := relayer.ValidateTransactionAndParseFulfillmentEvents(txHash)
	assert.NoError(t, err)
	assert.Equal(t, models.TransactionStatusFailed, result.TxStatus)
}

func TestValidateTransactionAndParseFulfillmentEvents_InvalidEvent(t *testing.T) {
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
	}

	mintControllerAddress := common.HexToAddress("0x1")
	mockMintController.EXPECT().Address().Return(mintControllerAddress)

	txHash := "0x1"
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(90),
		Logs: []*types.Log{
			{
				Address: common.HexToAddress("0x1"),
			},
		},
	}

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil)
	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))
	mockMintController.EXPECT().ParseFulfillment(mock.Anything).Return(&autogen.MintControllerFulfillment{}, assert.AnError)

	result, err := relayer.ValidateTransactionAndParseFulfillmentEvents(txHash)
	assert.NoError(t, err)
	assert.Equal(t, models.TransactionStatusInvalid, result.TxStatus)
	assert.Equal(t, uint64(10), result.Confirmations)
	assert.Empty(t, result.Events)
}

func TestValidateTransactionAndParseFulfillmentEvents(t *testing.T) {
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
	}

	mintControllerAddress := common.HexToAddress("0x1")
	mockMintController.EXPECT().Address().Return(mintControllerAddress)

	txHash := "0x1"
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(90),
		Logs: []*types.Log{
			{
				Address: common.HexToAddress("0x1"),
			},
		},
	}

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil)
	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))
	mockMintController.EXPECT().ParseFulfillment(mock.Anything).Return(&autogen.MintControllerFulfillment{}, nil)

	result, err := relayer.ValidateTransactionAndParseFulfillmentEvents(txHash)
	assert.NoError(t, err)
	assert.Equal(t, models.TransactionStatusConfirmed, result.TxStatus)
	assert.Equal(t, uint64(10), result.Confirmations)
	assert.NotEmpty(t, result.Events)
}

func TestRelayerUpdateTransaction_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "0x1",
	}

	update := bson.M{
		"status": models.TransactionStatusConfirmed,
	}

	mockDB.EXPECT().UpdateTransaction(txDoc.ID, update).Return(assert.AnError)

	success := relayer.UpdateTransaction(txDoc, update)
	assert.False(t, success)
}

func TestRelayerUpdateTransaction(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "0x1",
	}

	update := bson.M{
		"status": models.TransactionStatusConfirmed,
	}

	mockDB.EXPECT().UpdateTransaction(txDoc.ID, update).Return(nil)

	success := relayer.UpdateTransaction(txDoc, update)
	assert.True(t, success)
}

func TestRelayerConfirmFulfillmentTx_Nil(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	success := relayer.ConfirmFulfillmentTx(nil)
	assert.False(t, success)
}

func TestRelayerConfirmFulfillmentTx_Invalid(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	txHash := "0x1"

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(nil, assert.AnError)

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "0x1",
	}

	success := relayer.ConfirmFulfillmentTx(txDoc)
	assert.False(t, success)
}

func TestRelayerConfirmFulfillmentTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	mintControllerAddress := common.HexToAddress("0x1")
	mockMintController.EXPECT().Address().Return(mintControllerAddress)

	txHash := "0x1"
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(90),
		Logs: []*types.Log{
			{
				Address: common.HexToAddress("0x1"),
			},
		},
	}

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil)
	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))
	mockMintController.EXPECT().ParseFulfillment(mock.Anything).Return(&autogen.MintControllerFulfillment{}, nil)

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "0x1",
	}

	mockDB.EXPECT().UpdateTransaction(txDoc.ID, mock.Anything).Return(nil)

	success := relayer.ConfirmFulfillmentTx(txDoc)
	assert.True(t, success)
}

func TestRelayerConfirmMessagesForTx_Nil(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	success := relayer.ConfirmMessagesForTx(nil)
	assert.False(t, success)
}

func TestRelayerConfirmMessagesForTx_ValidationError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	txHash := "0x1"

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(nil, assert.AnError)

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "0x1",
	}

	success := relayer.ConfirmMessagesForTx(txDoc)
	assert.False(t, success)
}
func TestRelayerConfirmMessagesForTx_FailedTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	txHash := "0x1"
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusFailed,
		BlockNumber: big.NewInt(90),
		Logs: []*types.Log{
			{
				Address: common.HexToAddress("0x1"),
			},
		},
	}

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil)

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "0x1",
	}

	mockDB.EXPECT().UpdateTransaction(txDoc.ID, mock.Anything).Return(nil)

	success := relayer.ConfirmMessagesForTx(txDoc)
	assert.False(t, success)
}
func TestRelayerConfirmMessagesForTx_LockError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	mintControllerAddress := common.HexToAddress("0x1")
	mockMintController.EXPECT().Address().Return(mintControllerAddress)

	txHash := "0x1"
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(90),
		Logs: []*types.Log{
			{
				Address: common.HexToAddress("0x1"),
			},
		},
	}

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil)
	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))
	mockMintController.EXPECT().ParseFulfillment(mock.Anything).Return(&autogen.MintControllerFulfillment{}, nil)

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "0x1",
	}

	mockDB.EXPECT().LockWriteTransaction(txDoc).Return("lock-id", assert.AnError)

	success := relayer.ConfirmMessagesForTx(txDoc)
	assert.False(t, success)
}
func TestRelayerConfirmMessagesForTx_UpdateError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	mintControllerAddress := common.HexToAddress("0x1")
	mockMintController.EXPECT().Address().Return(mintControllerAddress)

	txHash := "0x1"
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(90),
		Logs: []*types.Log{
			{
				Address: common.HexToAddress("0x1"),
			},
		},
	}

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil)
	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))
	mockMintController.EXPECT().ParseFulfillment(mock.Anything).Return(&autogen.MintControllerFulfillment{}, nil)

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "0x1",
	}

	mockDB.EXPECT().LockWriteTransaction(txDoc).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().UpdateMessageByMessageID(mock.Anything, mock.Anything).Return(primitive.ObjectID{}, assert.AnError)

	success := relayer.ConfirmMessagesForTx(txDoc)
	assert.False(t, success)
}
func TestRelayerConfirmMessagesForTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	mintControllerAddress := common.HexToAddress("0x1")
	mockMintController.EXPECT().Address().Return(mintControllerAddress)

	txHash := "0x1"
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(90),
		Logs: []*types.Log{
			{
				Address: common.HexToAddress("0x1"),
			},
		},
	}

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil)
	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))
	mockMintController.EXPECT().ParseFulfillment(mock.Anything).Return(&autogen.MintControllerFulfillment{}, nil)

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "0x1",
	}

	mockDB.EXPECT().LockWriteTransaction(txDoc).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().UpdateMessageByMessageID(mock.Anything, mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateTransaction(txDoc.ID, mock.Anything).Return(nil)

	success := relayer.ConfirmMessagesForTx(txDoc)
	assert.True(t, success)
}

func TestRelayerSyncBlocks(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	startBlockHeight := uint64(1)
	endBlockHeight := uint64(100)

	iterator := clientMocks.NewMockMintControllerFulfillmentIterator(t)

	mockMintController.EXPECT().FilterFulfillment(&bind.FilterOpts{
		Start:   startBlockHeight,
		End:     &endBlockHeight,
		Context: context.Background(),
	}, mock.Anything).Return(iterator, nil)

	event := &autogen.MintControllerFulfillment{
		Raw: types.Log{
			TxHash: common.HexToHash("0x1"),
		},
	}

	iterator.EXPECT().Next().Return(true).Once()
	iterator.EXPECT().Event().Return(event).Once()
	iterator.EXPECT().Next().Return(false).Once()
	iterator.EXPECT().Error().Return(nil)
	iterator.EXPECT().Close().Return(nil)

	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))

	ethValidateTransactionByHash = func(client eth.EthereumClient, txHash string) (*ValidateTransactionByHashResult, error) {
		assert.Equal(t, mockEthClient, client)
		assert.Equal(t, common.HexToHash("0x1").Hex(), txHash)
		return &ValidateTransactionByHashResult{
			Receipt: &types.Receipt{Status: types.ReceiptStatusSuccessful},
			Tx:      &types.Transaction{},
		}, nil
	}

	mockDB.EXPECT().NewEthereumTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.Transaction{}, nil)
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, nil)

	success := relayer.SyncBlocks(startBlockHeight, endBlockHeight)
	assert.True(t, success)
}

func TestRelayerSyncBlocks_ErrorFiltering(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	startBlockHeight := uint64(1)
	endBlockHeight := uint64(100)

	iterator := clientMocks.NewMockMintControllerFulfillmentIterator(t)

	mockMintController.EXPECT().FilterFulfillment(&bind.FilterOpts{
		Start:   startBlockHeight,
		End:     &endBlockHeight,
		Context: context.Background(),
	}, mock.Anything).Return(iterator, assert.AnError)

	iterator.EXPECT().Close().Return(nil)

	success := relayer.SyncBlocks(startBlockHeight, endBlockHeight)
	assert.False(t, success)
}

func TestRelayerSyncBlocks_SecondError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	startBlockHeight := uint64(1)
	endBlockHeight := uint64(100)

	iterator := clientMocks.NewMockMintControllerFulfillmentIterator(t)

	mockMintController.EXPECT().FilterFulfillment(&bind.FilterOpts{
		Start:   startBlockHeight,
		End:     &endBlockHeight,
		Context: context.Background(),
	}, mock.Anything).Return(iterator, nil)

	event := &autogen.MintControllerFulfillment{
		Raw: types.Log{
			TxHash: common.HexToHash("0x1"),
		},
	}

	iterator.EXPECT().Next().Return(true).Twice()
	iterator.EXPECT().Event().Return(event).Once()
	iterator.EXPECT().Error().Return(nil).Once()
	iterator.EXPECT().Error().Return(assert.AnError)
	iterator.EXPECT().Close().Return(nil)

	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))

	ethValidateTransactionByHash = func(client eth.EthereumClient, txHash string) (*ValidateTransactionByHashResult, error) {
		assert.Equal(t, mockEthClient, client)
		assert.Equal(t, common.HexToHash("0x1").Hex(), txHash)
		return &ValidateTransactionByHashResult{
			Receipt: &types.Receipt{Status: types.ReceiptStatusSuccessful},
			Tx:      &types.Transaction{},
		}, nil
	}

	mockDB.EXPECT().NewEthereumTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.Transaction{}, nil)
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, nil)

	success := relayer.SyncBlocks(startBlockHeight, endBlockHeight)
	assert.False(t, success)
}

func TestRelayerSyncBlocks_EventNil(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	startBlockHeight := uint64(1)
	endBlockHeight := uint64(100)

	iterator := clientMocks.NewMockMintControllerFulfillmentIterator(t)

	mockMintController.EXPECT().FilterFulfillment(&bind.FilterOpts{
		Start:   startBlockHeight,
		End:     &endBlockHeight,
		Context: context.Background(),
	}, mock.Anything).Return(iterator, nil)

	iterator.EXPECT().Next().Return(true).Once()
	iterator.EXPECT().Event().Return(nil).Once()
	iterator.EXPECT().Next().Return(false).Once()
	iterator.EXPECT().Error().Return(nil)
	iterator.EXPECT().Close().Return(nil)

	success := relayer.SyncBlocks(startBlockHeight, endBlockHeight)
	assert.False(t, success)
}

func TestRelayerSyncBlocks_EventRemoved(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	startBlockHeight := uint64(1)
	endBlockHeight := uint64(100)

	iterator := clientMocks.NewMockMintControllerFulfillmentIterator(t)

	mockMintController.EXPECT().FilterFulfillment(&bind.FilterOpts{
		Start:   startBlockHeight,
		End:     &endBlockHeight,
		Context: context.Background(),
	}, mock.Anything).Return(iterator, nil)

	event := &autogen.MintControllerFulfillment{
		Raw: types.Log{
			TxHash:  common.HexToHash("0x1"),
			Removed: true,
		},
	}

	iterator.EXPECT().Next().Return(true).Once()
	iterator.EXPECT().Event().Return(event).Once()
	iterator.EXPECT().Next().Return(false).Once()
	iterator.EXPECT().Error().Return(nil)
	iterator.EXPECT().Close().Return(nil)

	success := relayer.SyncBlocks(startBlockHeight, endBlockHeight)
	assert.True(t, success)
}

func TestRelayerSyncNewBlock_NoBlocks(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	startBlockHeight := uint64(100)
	endBlockHeight := uint64(100)

	relayer := &EthMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		startBlockHeight:   startBlockHeight,
		currentBlockHeight: endBlockHeight,
	}

	success := relayer.SyncNewBlocks()
	assert.True(t, success)
}

func TestRelayerSyncNewBlock_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	startBlockHeight := uint64(1)
	endBlockHeight := uint64(100)

	relayer := &EthMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		startBlockHeight:   startBlockHeight,
		currentBlockHeight: endBlockHeight,
	}

	iterator := clientMocks.NewMockMintControllerFulfillmentIterator(t)

	mockMintController.EXPECT().FilterFulfillment(&bind.FilterOpts{
		Start:   startBlockHeight,
		End:     &endBlockHeight,
		Context: context.Background(),
	}, mock.Anything).Return(iterator, nil)

	iterator.EXPECT().Next().Return(false).Once()
	iterator.EXPECT().Error().Return(assert.AnError)
	iterator.EXPECT().Close().Return(nil)

	success := relayer.SyncNewBlocks()
	assert.False(t, success)
}

func TestRelayerSyncNewBlock(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	startBlockHeight := uint64(1)
	endBlockHeight := uint64(100)

	relayer := &EthMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		startBlockHeight:   startBlockHeight,
		currentBlockHeight: endBlockHeight,
	}

	iterator := clientMocks.NewMockMintControllerFulfillmentIterator(t)

	mockMintController.EXPECT().FilterFulfillment(&bind.FilterOpts{
		Start:   startBlockHeight,
		End:     &endBlockHeight,
		Context: context.Background(),
	}, mock.Anything).Return(iterator, nil)

	event := &autogen.MintControllerFulfillment{
		Raw: types.Log{
			TxHash: common.HexToHash("0x1"),
		},
	}

	iterator.EXPECT().Next().Return(true).Once()
	iterator.EXPECT().Event().Return(event).Once()
	iterator.EXPECT().Next().Return(false).Once()
	iterator.EXPECT().Error().Return(nil)
	iterator.EXPECT().Close().Return(nil)

	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))

	ethValidateTransactionByHash = func(client eth.EthereumClient, txHash string) (*ValidateTransactionByHashResult, error) {
		assert.Equal(t, mockEthClient, client)
		assert.Equal(t, common.HexToHash("0x1").Hex(), txHash)
		return &ValidateTransactionByHashResult{
			Receipt: &types.Receipt{Status: types.ReceiptStatusSuccessful},
			Tx:      &types.Transaction{},
		}, nil
	}

	mockDB.EXPECT().NewEthereumTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.Transaction{}, nil)
	mockDB.EXPECT().InsertTransaction(mock.Anything).Return(primitive.ObjectID{}, nil)

	success := relayer.SyncNewBlocks()
	assert.True(t, success)
}

func TestRelayerSyncNewBlock_EthQueryMax(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	startBlockHeight := uint64(1)
	endBlockHeight := uint64(100) + eth.MaxQueryBlocks

	relayer := &EthMessageRelayerRunnable{
		db:                 mockDB,
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		startBlockHeight:   startBlockHeight,
		currentBlockHeight: endBlockHeight,
	}

	iterator := clientMocks.NewMockMintControllerFulfillmentIterator(t)

	endBlock := startBlockHeight + eth.MaxQueryBlocks

	mockMintController.EXPECT().FilterFulfillment(&bind.FilterOpts{
		Start:   startBlockHeight,
		End:     &endBlock,
		Context: context.Background(),
	}, mock.Anything).Return(iterator, nil)

	mockMintController.EXPECT().FilterFulfillment(&bind.FilterOpts{
		Start:   endBlock,
		End:     &endBlockHeight,
		Context: context.Background(),
	}, mock.Anything).Return(iterator, nil)

	iterator.EXPECT().Next().Return(false)
	iterator.EXPECT().Error().Return(nil)
	iterator.EXPECT().Close().Return(nil)

	success := relayer.SyncNewBlocks()
	assert.True(t, success)
}

func TestConfirmFulfillmentTxs_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	mintControllerAddress := common.HexToAddress("0x1")
	mockMintController.EXPECT().Address().Return(mintControllerAddress)

	mockDB.EXPECT().GetPendingTransactionsTo(mock.Anything, mintControllerAddress.Bytes()).Return(nil, assert.AnError)

	success := relayer.ConfirmFulfillmentTxs()
	assert.False(t, success)
}

func TestConfirmFulfillmentTxs(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	mintControllerAddress := common.HexToAddress("0x1")
	mockMintController.EXPECT().Address().Return(mintControllerAddress)

	txHash := "0x1"
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(90),
		Logs: []*types.Log{
			{
				Address: common.HexToAddress("0x1"),
			},
		},
	}

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil)
	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))
	mockMintController.EXPECT().ParseFulfillment(mock.Anything).Return(&autogen.MintControllerFulfillment{}, nil)

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "0x1",
	}

	mockDB.EXPECT().GetPendingTransactionsTo(mock.Anything, mintControllerAddress.Bytes()).Return([]models.Transaction{*txDoc}, nil)
	mockDB.EXPECT().UpdateTransaction(txDoc.ID, mock.Anything).Return(nil)

	success := relayer.ConfirmFulfillmentTxs()
	assert.True(t, success)
}

func TestRelayerConfirmMessages_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	mintControllerAddress := common.HexToAddress("0x1")
	mockMintController.EXPECT().Address().Return(mintControllerAddress)
	mockDB.EXPECT().GetConfirmedTransactionsTo(mock.Anything, mintControllerAddress.Bytes()).Return(nil, assert.AnError)

	success := relayer.ConfirmMessages()
	assert.False(t, success)
}

func TestRelayerConfirmMessages(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:             mockEthClient,
		mintController:     mockMintController,
		logger:             logger,
		currentBlockHeight: 100,
		confirmations:      10,
		db:                 mockDB,
	}

	mintControllerAddress := common.HexToAddress("0x1")
	mockMintController.EXPECT().Address().Return(mintControllerAddress)

	txHash := "0x1"
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(90),
		Logs: []*types.Log{
			{
				Address: common.HexToAddress("0x1"),
			},
		},
	}

	mockEthClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil)
	mockMintController.EXPECT().Address().Return(common.HexToAddress("0x1"))
	mockMintController.EXPECT().ParseFulfillment(mock.Anything).Return(&autogen.MintControllerFulfillment{}, nil)

	txDoc := &models.Transaction{
		ID:   &primitive.ObjectID{},
		Hash: "0x1",
	}

	mockDB.EXPECT().LockWriteTransaction(txDoc).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().UpdateMessageByMessageID(mock.Anything, mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateTransaction(txDoc.ID, mock.Anything).Return(nil)

	mockDB.EXPECT().GetConfirmedTransactionsTo(mock.Anything, mintControllerAddress.Bytes()).Return([]models.Transaction{*txDoc}, nil)

	success := relayer.ConfirmMessages()
	assert.True(t, success)
}

func TestRelayerInitStartBlockHeight(t *testing.T) {
	logger := log.New().WithField("test", "relayer")
	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

	relayer := &EthMessageRelayerRunnable{
		logger:             logger,
		startBlockHeight:   0,
		currentBlockHeight: 200,
	}

	relayer.InitStartBlockHeight(lastHealth)

	assert.Equal(t, uint64(100), relayer.startBlockHeight)

	relayer = &EthMessageRelayerRunnable{
		logger:             logger,
		startBlockHeight:   0,
		currentBlockHeight: 200,
	}
	lastHealth = nil
	relayer.InitStartBlockHeight(lastHealth)
	assert.Equal(t, uint64(200), relayer.startBlockHeight)

	relayer = &EthMessageRelayerRunnable{
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
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := log.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		client:           mockEthClient,
		mintController:   mockMintController,
		logger:           logger,
		startBlockHeight: 100,
		confirmations:    10,
		db:               mockDB,
	}

	mintControllerAddress := common.HexToAddress("0x1")
	mockMintController.EXPECT().Address().Return(mintControllerAddress)

	mockEthClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	mockDB.EXPECT().GetPendingTransactionsTo(mock.Anything, mock.Anything).Return([]models.Transaction{}, nil)
	mockDB.EXPECT().GetConfirmedTransactionsTo(mock.Anything, mock.Anything).Return([]models.Transaction{}, nil)

	relayer.Run()
}

func TestNewMessageRelayer(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)

	mintControllerMap := map[uint32][]byte{
		1: ethcommon.FromHex("0x01"),
		2: ethcommon.FromHex("0x02"),
	}

	lastHealth := &models.RunnerServiceStatus{BlockHeight: 50}

	config := models.EthereumNetworkConfig{
		ChainID:               1,
		ChainName:             "test",
		StartBlockHeight:      1,
		Confirmations:         10,
		MintControllerAddress: ethcommon.BytesToAddress([]byte("mintController")).Hex(),
		MessageRelayer: models.ServiceConfig{
			Enabled: true,
		},
	}

	ethNewClient = func(models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		return mockClient, nil
	}

	ethNewMintControllerContract = func(ethcommon.Address, bind.ContractBackend) (eth.MintControllerContract, error) {
		return mockMintController, nil
	}

	dbNewDB = func() db.DB {
		return mockDB
	}

	defer func() {
		ethNewClient = eth.NewClient
		ethNewMintControllerContract = eth.NewMintControllerContract
		dbNewDB = db.NewDB
	}()

	mockClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)

	mockClient.EXPECT().GetClient().Return(nil)

	runnable := NewMessageRelayer(config, mintControllerMap, lastHealth)

	assert.NotNil(t, runnable)

	monitor, ok := runnable.(*EthMessageRelayerRunnable)
	assert.True(t, ok)
	assert.Equal(t, mockDB, monitor.db)
	assert.Equal(t, mockClient, monitor.client)
	assert.Equal(t, mockMintController, monitor.mintController)
	assert.Equal(t, mintControllerMap, monitor.mintControllerMap)
	assert.Equal(t, uint64(100), monitor.currentBlockHeight)
	assert.Equal(t, uint64(50), monitor.startBlockHeight)

}

func TestNewMessageRelayerFailures(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)

	mintControllerMap := map[uint32][]byte{
		1: ethcommon.FromHex("0x01"),
		2: ethcommon.FromHex("0x02"),
	}

	lastHealth := &models.RunnerServiceStatus{BlockHeight: 50}

	config := models.EthereumNetworkConfig{
		ChainID:               1,
		ChainName:             "test",
		StartBlockHeight:      1,
		Confirmations:         10,
		MintControllerAddress: ethcommon.BytesToAddress([]byte("mintController")).Hex(),
		MessageRelayer: models.ServiceConfig{
			Enabled: true,
		},
	}

	ethNewClient = func(models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		return mockClient, nil
	}

	ethNewMintControllerContract = func(ethcommon.Address, bind.ContractBackend) (eth.MintControllerContract, error) {
		return mockMintController, nil
	}

	dbNewDB = func() db.DB {
		return mockDB
	}

	defer func() {
		ethNewClient = eth.NewClient
		ethNewMailboxContract = eth.NewMailboxContract
		dbNewDB = db.NewDB
	}()

	mockClient.EXPECT().GetClient().Return(nil)

	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	t.Run("Disabled", func(t *testing.T) {
		config.MessageRelayer.Enabled = false

		assert.Panics(t, func() {
			NewMessageRelayer(config, mintControllerMap, lastHealth)
		})

		config.MessageRelayer.Enabled = true
	})

	t.Run("ClientError", func(t *testing.T) {

		ethNewClient = func(models.EthereumNetworkConfig) (eth.EthereumClient, error) {
			return nil, assert.AnError
		}

		assert.Panics(t, func() {
			NewMessageRelayer(config, mintControllerMap, lastHealth)
		})

		ethNewClient = func(models.EthereumNetworkConfig) (eth.EthereumClient, error) {
			return mockClient, nil
		}

	})

	t.Run("MintControllerError", func(t *testing.T) {

		ethNewMintControllerContract = func(ethcommon.Address, bind.ContractBackend) (eth.MintControllerContract, error) {
			return nil, assert.AnError
		}

		assert.Panics(t, func() {
			NewMessageRelayer(config, mintControllerMap, lastHealth)
		})

		ethNewMintControllerContract = func(ethcommon.Address, bind.ContractBackend) (eth.MintControllerContract, error) {
			return mockMintController, nil
		}

	})

}
