package ethereum

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	clientMocks "github.com/dan13ram/wpokt-oracle/ethereum/client/mocks"
	"github.com/dan13ram/wpokt-oracle/models"
)

func TestRelayerUpdateCurrentBlockHeight(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:     mockDB,
		client: mockEthClient,
		logger: logger,
	}

	mockEthClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)

	relayer.UpdateCurrentBlockHeight()

	assert.Equal(t, uint64(100), relayer.currentBlockHeight)
}

func TestCreateTxForFulfillmentEvent(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := logrus.New().WithField("test", "relayer")

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

func TestValidateTransactionAndParseFulfillmentEvents(t *testing.T) {
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := logrus.New().WithField("test", "relayer")

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

func TestRelayerUpdateTransaction(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := logrus.New().WithField("test", "relayer")

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

func TestRelayerConfirmFulfillmentTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := logrus.New().WithField("test", "relayer")

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

func TestRelayerConfirmMessagesForTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := logrus.New().WithField("test", "relayer")

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

	relayer.ConfirmMessagesForTx(txDoc)
}

func TestRelayerSyncBlocks(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := logrus.New().WithField("test", "relayer")

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

func TestRelayerSyncNewBlock(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := logrus.New().WithField("test", "relayer")

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

func TestConfirmFulfillmentTxs(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := logrus.New().WithField("test", "relayer")

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

	relayer.ConfirmFulfillmentTxs()
}

func TestRelayerConfirmMessages(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := logrus.New().WithField("test", "relayer")

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

	relayer.ConfirmMessages()
}

func TestRelayerInitStartBlockHeight(t *testing.T) {
	logger := logrus.New().WithField("test", "relayer")
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

/*
func TestRelayerRun(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	logger := logrus.New().WithField("test", "relayer")

	relayer := &EthMessageRelayerRunnable{
		db:             mockDB,
		client:         mockEthClient,
		mintController: mockMintController,
		logger:         logger,
	}

	mockEthClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	mockDB.EXPECT().GetPendingTransactionsTo(mock.Anything, mock.Anything).Return([]models.Transaction{}, nil)
	mockDB.EXPECT().GetConfirmedTransactionsTo(mock.Anything, mock.Anything).Return([]models.Transaction{}, nil)

	relayer.Run()
}
*/
