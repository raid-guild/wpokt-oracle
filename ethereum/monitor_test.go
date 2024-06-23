package ethereum

import (
	"context"
	"math/big"
	"testing"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	clientMocks "github.com/dan13ram/wpokt-oracle/ethereum/client/mocks"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMonitorUpdateCurrentBlockHeight(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "monitor")

	monitor := &EthMessageMonitorRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)

	monitor.UpdateCurrentBlockHeight()

	assert.Equal(t, uint64(100), monitor.currentBlockHeight)
}

func TestUpdateTransaction(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := logrus.New().WithField("test", "monitor")

	tx := &models.Transaction{ID: &primitive.ObjectID{}}
	update := bson.M{"status": models.TransactionStatusConfirmed}

	monitor := &EthMessageMonitorRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateTransaction(tx.ID, update).Return(nil)

	result := monitor.UpdateTransaction(tx, update)

	assert.True(t, result)
}

func createValidEvent(t *testing.T) (*autogen.MailboxDispatch, models.MessageContent) {
	mintControllerOne := ethcommon.HexToAddress("0x0301")
	mintControllerTwo := ethcommon.HexToAddress("0x0302")

	event := &autogen.MailboxDispatch{
		Sender:      mintControllerOne,
		Recipient:   ethcommon.BytesToHash(mintControllerTwo.Bytes()),
		Destination: 2,
		Message:     []byte{},
		Raw: types.Log{
			TxHash: ethcommon.HexToHash("0x01"),
		},
	}

	messageContent := models.MessageContent{
		OriginDomain:      1,
		DestinationDomain: 2,
		Version:           common.HyperlaneVersion,
		Sender:            mintControllerOne.Hex(),
		Recipient:         mintControllerTwo.Hex(),
		MessageBody: models.MessageBody{
			Amount:           "100",
			SenderAddress:    common.HexToAddress("0x01").Hex(),
			RecipientAddress: common.HexToAddress("0x02").Hex(),
		},
	}

	bytes, err := messageContent.EncodeToBytes()
	assert.NoError(t, err)
	event.Message = bytes

	return event, messageContent
}

func TestIsValidEvent(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := logrus.New().WithField("test", "monitor")

	mintControllerOne := ethcommon.HexToAddress("0x0301")
	mintControllerTwo := ethcommon.HexToAddress("0x0302")

	mintControllerMap := map[uint32][]byte{
		1: mintControllerOne.Bytes(),
		2: mintControllerTwo.Bytes(),
	}

	monitor := &EthMessageMonitorRunnable{
		db:                mockDB,
		logger:            logger,
		mintControllerMap: mintControllerMap,
		chain: models.Chain{
			ChainDomain: 1,
		},
	}

	event, _ := createValidEvent(t)

	err := monitor.IsValidEvent(event)
	assert.NoError(t, err)
}

func TestCreateTxForDispatchEvent(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "monitor")

	mintControllerOne := ethcommon.HexToAddress("0x0301")
	mintControllerTwo := ethcommon.HexToAddress("0x0302")

	mintControllerMap := map[uint32][]byte{
		1: mintControllerOne.Bytes(),
		2: mintControllerTwo.Bytes(),
	}

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
		chain: models.Chain{
			ChainDomain: 1,
		},
		mailbox: mailbox,
	}

	event, _ := createValidEvent(t)

	mailbox.EXPECT().Address().Return(ethcommon.Address{})

	tx := models.Transaction{ID: &primitive.ObjectID{}, Hash: "0x01"}
	mockDB.EXPECT().NewEthereumTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, models.TransactionStatusPending).Return(tx, nil)
	mockDB.EXPECT().InsertTransaction(tx).Return(primitive.ObjectID{}, nil)

	ethValidateTransactionByHash = func(client eth.EthereumClient, hash string) (*ValidateTransactionByHashResult, error) {
		result := &ValidateTransactionByHashResult{
			Tx:      &types.Transaction{},
			Receipt: &types.Receipt{Status: types.ReceiptStatusSuccessful},
		}
		assert.Equal(t, ethcommon.HexToHash("0x01").Hex(), hash)
		assert.Equal(t, mockClient, client)
		return result, nil
	}

	result := monitor.CreateTxForDispatchEvent(event)

	assert.True(t, result)
}

func TestValidateTransactionAndParseDispatchEvents(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	receipt := &types.Receipt{
		BlockNumber: big.NewInt(100),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{},
	}

	mockClient.EXPECT().GetTransactionReceipt("0x01").Return(receipt, nil)

	result, err := monitor.ValidateTransactionAndParseDispatchEvents("0x01")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.TransactionStatusInvalid, result.TxStatus)
}

func TestConfirmTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	tx := &models.Transaction{ID: &primitive.ObjectID{}, Hash: "0x01"}

	receipt := &types.Receipt{
		BlockNumber: big.NewInt(100),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{},
	}

	mockClient.EXPECT().GetTransactionReceipt("0x01").Return(receipt, nil)

	mockDB.EXPECT().UpdateTransaction(tx.ID, mock.Anything).Return(nil)

	result := monitor.ConfirmTx(tx)

	assert.True(t, result)
}

func TestCreateMessagesForTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	mailbox.EXPECT().Address().Return(ethcommon.Address{})

	event, content := createValidEvent(t)

	mailbox.EXPECT().ParseDispatch(mock.Anything).Return(event, nil)

	tx := &models.Transaction{ID: &primitive.ObjectID{}, Hash: "0x01"}

	receipt := &types.Receipt{
		BlockNumber: big.NewInt(100),
		Status:      types.ReceiptStatusSuccessful,
		Logs: []*types.Log{
			{
				Address: ethcommon.Address{},
			},
		},
	}

	mockClient.EXPECT().GetTransactionReceipt("0x01").Return(receipt, nil)

	mockDB.EXPECT().LockWriteTransaction(tx).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().InsertMessage(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateTransaction(tx.ID, mock.Anything).Return(nil)

	mockDB.EXPECT().NewMessage(tx, content, models.MessageStatusPending).Return(models.Message{}, nil)
	mockDB.EXPECT().InsertMessage(mock.Anything).Return(primitive.ObjectID{}, nil)

	result := monitor.CreateMessagesForTx(tx)

	assert.True(t, result)
}

func TestMonitorSyncBlocks(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "monitor")

	mintControllerMap := map[uint32][]byte{
		1: ethcommon.FromHex("0x01"),
	}

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
		chain: models.Chain{
			ChainDomain: 1,
		},
		mailbox: mailbox,
	}

	mintControllerAddress := ethcommon.BytesToAddress(mintControllerMap[1])

	startBlock := uint64(1)
	endBlock := uint64(100)

	filter := &bind.FilterOpts{
		Start:   startBlock,
		End:     &endBlock,
		Context: context.Background(),
	}

	iterator := clientMocks.NewMockMailboxDispatchIterator(t)

	mailbox.EXPECT().FilterDispatch(filter, []ethcommon.Address{mintControllerAddress}, []uint32{}, [][32]byte{}).Return(iterator, nil)

	iterator.EXPECT().Next().Return(false)
	iterator.EXPECT().Error().Return(nil)
	iterator.EXPECT().Close().Return(nil)

	result := monitor.SyncBlocks(startBlock, endBlock)

	assert.True(t, result)
}

func TestMonitorSyncNewBlocks(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "monitor")

	mintControllerMap := map[uint32][]byte{
		1: ethcommon.FromHex("0x01"),
	}

	startBlock := uint64(1)
	endBlock := uint64(100)

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:                mockDB,
		client:            mockClient,
		logger:            logger,
		mintControllerMap: mintControllerMap,
		chain: models.Chain{
			ChainDomain: 1,
		},
		mailbox:            mailbox,
		currentBlockHeight: endBlock,
		startBlockHeight:   startBlock,
	}

	filter := &bind.FilterOpts{
		Start:   startBlock,
		End:     &endBlock,
		Context: context.Background(),
	}

	iterator := clientMocks.NewMockMailboxDispatchIterator(t)

	mintControllerAddress := ethcommon.BytesToAddress(mintControllerMap[1])
	mailbox.EXPECT().FilterDispatch(filter, []ethcommon.Address{mintControllerAddress}, []uint32{}, [][32]byte{}).Return(iterator, nil)

	iterator.EXPECT().Next().Return(false)
	iterator.EXPECT().Error().Return(nil)
	iterator.EXPECT().Close().Return(nil)

	result := monitor.SyncNewBlocks()

	assert.True(t, result)
}

func TestConfirmDispatchTxs(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	tx := models.Transaction{ID: &primitive.ObjectID{}, Hash: "0x01"}

	mockDB.EXPECT().GetPendingTransactionsTo(mock.Anything, mock.Anything).Return([]models.Transaction{tx}, nil)
	mockClient.EXPECT().GetTransactionReceipt("0x01").Return(&types.Receipt{
		BlockNumber: big.NewInt(100),
		Status:      types.ReceiptStatusSuccessful}, nil)
	mockDB.EXPECT().UpdateTransaction(tx.ID, mock.Anything).Return(nil)
	mailbox.EXPECT().Address().Return(ethcommon.Address{})

	result := monitor.ConfirmDispatchTxs()

	assert.True(t, result)
}

func TestCreateMessagesForTxs(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	tx := models.Transaction{ID: &primitive.ObjectID{}, Hash: "0x01"}

	mockDB.EXPECT().GetConfirmedTransactionsTo(mock.Anything, mock.Anything).Return([]models.Transaction{tx}, nil)

	mailbox.EXPECT().Address().Return(ethcommon.Address{})

	event, content := createValidEvent(t)

	mailbox.EXPECT().ParseDispatch(mock.Anything).Return(event, nil)

	receipt := &types.Receipt{
		BlockNumber: big.NewInt(100),
		Status:      types.ReceiptStatusSuccessful,
		Logs: []*types.Log{
			{
				Address: ethcommon.Address{},
			},
		},
	}

	mockClient.EXPECT().GetTransactionReceipt("0x01").Return(receipt, nil)

	mockDB.EXPECT().LockWriteTransaction(&tx).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().InsertMessage(mock.Anything).Return(primitive.ObjectID{}, nil)
	mockDB.EXPECT().UpdateTransaction(tx.ID, mock.Anything).Return(nil)

	mockDB.EXPECT().NewMessage(&tx, content, models.MessageStatusPending).Return(models.Message{}, nil)
	mockDB.EXPECT().InsertMessage(mock.Anything).Return(primitive.ObjectID{}, nil)

	result := monitor.CreateMessagesForTxs()

	assert.True(t, result)
}

func TestMonitorInitStartBlockHeight(t *testing.T) {
	logger := logrus.New().WithField("test", "monitor")
	lastHealth := &models.RunnerServiceStatus{BlockHeight: 100}

	monitor := &EthMessageMonitorRunnable{
		logger:             logger,
		startBlockHeight:   0,
		currentBlockHeight: 200,
	}

	monitor.InitStartBlockHeight(lastHealth)

	assert.Equal(t, uint64(100), monitor.startBlockHeight)

	monitor = &EthMessageMonitorRunnable{
		logger:             logger,
		startBlockHeight:   0,
		currentBlockHeight: 200,
	}
	lastHealth = nil
	monitor.InitStartBlockHeight(lastHealth)
	assert.Equal(t, uint64(200), monitor.startBlockHeight)

	monitor = &EthMessageMonitorRunnable{
		logger:             logger,
		startBlockHeight:   0,
		currentBlockHeight: 200,
	}
	lastHealth = &models.RunnerServiceStatus{BlockHeight: 300}
	monitor.InitStartBlockHeight(lastHealth)
	assert.Equal(t, uint64(200), monitor.startBlockHeight)
}

func TestMonitorRun(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	mockClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	mockDB.EXPECT().GetPendingTransactionsTo(mock.Anything, mock.Anything).Return([]models.Transaction{}, nil)
	mockDB.EXPECT().GetConfirmedTransactionsTo(mock.Anything, mock.Anything).Return([]models.Transaction{}, nil)
	mailbox.EXPECT().Address().Return(ethcommon.Address{})

	monitor.Run()

	mockClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}
