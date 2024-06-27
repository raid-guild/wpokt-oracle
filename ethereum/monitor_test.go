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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	log "github.com/sirupsen/logrus"
)

func TestMonitorHeight(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	monitor := &EthMessageMonitorRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
	}

	height := monitor.Height()

	assert.Equal(t, uint64(100), height)
}

func TestMonitorUpdateCurrentBlockHeight(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	monitor := &EthMessageMonitorRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)

	monitor.UpdateCurrentBlockHeight()

	assert.Equal(t, uint64(100), monitor.currentBlockHeight)
}

func TestMonitorUpdateCurrentBlockHeight_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	monitor := &EthMessageMonitorRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetBlockHeight().Return(uint64(100), assert.AnError)

	monitor.UpdateCurrentBlockHeight()

	assert.Equal(t, uint64(0), monitor.currentBlockHeight)
}

func TestMonitorUpdateTransaction(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

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

func TestMonitorUpdateTransaction_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

	tx := &models.Transaction{ID: &primitive.ObjectID{}}
	update := bson.M{"status": models.TransactionStatusConfirmed}

	monitor := &EthMessageMonitorRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateTransaction(tx.ID, update).Return(assert.AnError)

	result := monitor.UpdateTransaction(tx, update)

	assert.False(t, result)
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
	logger := log.New().WithField("test", "monitor")

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

func TestIsValidEvent_Nil(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

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

	err := monitor.IsValidEvent(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event is nil")
}

func TestIsValidEvent_NoMintController(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

	mintControllerTwo := ethcommon.HexToAddress("0x0302")

	mintControllerMap := map[uint32][]byte{
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
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mint controller not found for chain domain")
}

func TestIsValidEvent_NoDestMintController(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

	mintControllerOne := ethcommon.HexToAddress("0x0301")

	mintControllerMap := map[uint32][]byte{
		1: mintControllerOne.Bytes(),
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
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mint controller not found for destination domain")
}

func TestIsValidEvent_InvalidSender(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

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
	event.Sender = ethcommon.HexToAddress("0x0303")

	err := monitor.IsValidEvent(event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sender does not match mint controller for chain domain")
}

func TestIsValidEvent_InvalidRecipient(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

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
	event.Recipient = ethcommon.HexToHash("0x0303")

	err := monitor.IsValidEvent(event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recipient does not match mint controller for destination domain")
}

func TestIsValidEvent_MessageContent_InvalidMessage(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

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
	event.Message = []byte{0x01}

	err := monitor.IsValidEvent(event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error decoding message content")
}

func TestIsValidEvent_MessageContent_InvalidDomain(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

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

	messageContent := models.MessageContent{
		OriginDomain:      3,
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

	err = monitor.IsValidEvent(event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid origin domain")
}

func TestIsValidEvent_MessageContent_InvalidDestDomain(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

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

	messageContent := models.MessageContent{
		OriginDomain:      1,
		DestinationDomain: 3,
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

	err = monitor.IsValidEvent(event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid destination domain")
}

func TestIsValidEvent_MessageContent_InvalidVersion(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

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

	messageContent := models.MessageContent{
		OriginDomain:      1,
		DestinationDomain: 2,
		Version:           0,
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

	err = monitor.IsValidEvent(event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid version")
}

func TestIsValidEvent_MessageContent_InvalidSender(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

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

	messageContent := models.MessageContent{
		OriginDomain:      1,
		DestinationDomain: 2,
		Version:           common.HyperlaneVersion,
		Sender:            ethcommon.HexToAddress("0x03").Hex(),
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

	err = monitor.IsValidEvent(event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sender")
}

func TestIsValidEvent_MessageContent_InvalidRecipient(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "monitor")

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

	messageContent := models.MessageContent{
		OriginDomain:      1,
		DestinationDomain: 2,
		Version:           common.HyperlaneVersion,
		Sender:            mintControllerOne.Hex(),
		Recipient:         ethcommon.HexToAddress("0x03").Hex(),
		MessageBody: models.MessageBody{
			Amount:           "100",
			SenderAddress:    common.HexToAddress("0x01").Hex(),
			RecipientAddress: common.HexToAddress("0x02").Hex(),
		},
	}

	bytes, err := messageContent.EncodeToBytes()
	assert.NoError(t, err)
	event.Message = bytes

	err = monitor.IsValidEvent(event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid recipient")
}

func TestCreateTxForDispatchEvent(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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

func TestCreateTxForDispatchEvent_InvalidEvent(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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
	event.Message = []byte{0x01}

	result := monitor.CreateTxForDispatchEvent(event)

	assert.False(t, result)
}

func TestCreateTxForDispatchEvent_NewFailed(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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
	mockDB.EXPECT().NewEthereumTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything, models.TransactionStatusPending).Return(tx, assert.AnError)

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

	assert.False(t, result)
}

func TestCreateTxForDispatchEvent_InsertFailed(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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
	mockDB.EXPECT().InsertTransaction(tx).Return(primitive.ObjectID{}, assert.AnError)

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

	assert.False(t, result)
}

func TestCreateTxForDispatchEvent_ValidateFailed(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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

	ethValidateTransactionByHash = func(client eth.EthereumClient, hash string) (*ValidateTransactionByHashResult, error) {
		assert.Equal(t, ethcommon.HexToHash("0x01").Hex(), hash)
		assert.Equal(t, mockClient, client)
		return nil, assert.AnError
	}

	result := monitor.CreateTxForDispatchEvent(event)

	assert.False(t, result)
}

func TestValidateTransactionAndParseDispatchEvents_ClientError(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	mockClient.EXPECT().GetTransactionReceipt("0x01").Return(nil, assert.AnError)

	result, err := monitor.ValidateTransactionAndParseDispatchEvents("0x01")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestValidateTransactionAndParseDispatchEvents_Failed(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	receipt := &types.Receipt{
		BlockNumber: big.NewInt(100),
		Status:      types.ReceiptStatusFailed,
		Logs:        []*types.Log{},
	}

	mockClient.EXPECT().GetTransactionReceipt("0x01").Return(receipt, nil)

	result, err := monitor.ValidateTransactionAndParseDispatchEvents("0x01")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.TransactionStatusFailed, result.TxStatus)
}

func TestValidateTransactionAndParseDispatchEvents(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	mailbox.EXPECT().ParseDispatch(mock.Anything).Return(nil, assert.AnError)
	mailbox.EXPECT().Address().Return(ethcommon.Address{})
	monitor := &EthMessageMonitorRunnable{
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	receipt := &types.Receipt{
		BlockNumber: big.NewInt(100),
		Status:      types.ReceiptStatusSuccessful,
		Logs: []*types.Log{{
			Address: ethcommon.Address{},
		}},
	}

	mockClient.EXPECT().GetTransactionReceipt("0x01").Return(receipt, nil)

	result, err := monitor.ValidateTransactionAndParseDispatchEvents("0x01")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.TransactionStatusInvalid, result.TxStatus)
}

func TestConfirmTx_Nil(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	result := monitor.ConfirmTx(nil)

	assert.False(t, result)
}

func TestConfirmTx_Failed(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	tx := &models.Transaction{ID: &primitive.ObjectID{}, Hash: "0x01"}

	mockClient.EXPECT().GetTransactionReceipt("0x01").Return(nil, assert.AnError)

	result := monitor.ConfirmTx(tx)

	assert.False(t, result)
}

func TestConfirmTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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

func TestCreateMessagesForTx_Nil(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	result := monitor.CreateMessagesForTx(nil)

	assert.False(t, result)
}

func TestCreateMessagesForTx_Invalid(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	mailbox.EXPECT().Address().Return(ethcommon.Address{})

	event, _ := createValidEvent(t)

	mailbox.EXPECT().ParseDispatch(mock.Anything).Return(event, assert.AnError)

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
	mockDB.EXPECT().UpdateTransaction(tx.ID, mock.Anything).Return(nil)

	result := monitor.CreateMessagesForTx(tx)

	assert.False(t, result)
}

func TestCreateMessagesForTx_ErrorValidating(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	mockClient.EXPECT().GetTransactionReceipt("0x01").Return(nil, assert.AnError)

	tx := &models.Transaction{ID: &primitive.ObjectID{}, Hash: "0x01"}
	result := monitor.CreateMessagesForTx(tx)

	assert.False(t, result)
}

func TestCreateMessagesForTx_LockError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	mailbox.EXPECT().Address().Return(ethcommon.Address{})

	event, _ := createValidEvent(t)

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

	mockDB.EXPECT().LockWriteTransaction(tx).Return("lock-id", assert.AnError)

	result := monitor.CreateMessagesForTx(tx)

	assert.False(t, result)
}

func TestCreateMessagesForTx_NewError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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

	mockDB.EXPECT().NewMessage(tx, content, models.MessageStatusPending).Return(models.Message{}, assert.AnError)

	result := monitor.CreateMessagesForTx(tx)

	assert.False(t, result)
}

func TestCreateMessagesForTx_InsertError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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

	mockDB.EXPECT().NewMessage(tx, content, models.MessageStatusPending).Return(models.Message{}, nil)
	mockDB.EXPECT().InsertMessage(mock.Anything).Return(primitive.ObjectID{}, assert.AnError)

	result := monitor.CreateMessagesForTx(tx)

	assert.False(t, result)
}

func TestCreateMessagesForTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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
	logger := log.New().WithField("test", "monitor")

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

func TestMonitorSyncBlocks_SingleValidEvent(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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

	event, _ := createValidEvent(t)

	iterator.EXPECT().Next().Return(true).Once()
	iterator.EXPECT().Next().Return(false)
	iterator.EXPECT().Event().Return(event)
	iterator.EXPECT().Error().Return(nil)
	iterator.EXPECT().Close().Return(nil)

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

	result := monitor.SyncBlocks(startBlock, endBlock)

	assert.True(t, result)
}

func TestMonitorSyncBlocks_SingleValidEvent_SecondError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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

	event, _ := createValidEvent(t)

	iterator.EXPECT().Next().Return(true).Twice()
	iterator.EXPECT().Event().Return(event)
	iterator.EXPECT().Error().Return(nil).Once()
	iterator.EXPECT().Error().Return(assert.AnError)
	iterator.EXPECT().Close().Return(nil)

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

	result := monitor.SyncBlocks(startBlock, endBlock)

	assert.False(t, result)
}

func TestMonitorSyncBlocks_SingleInvalidEvent(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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

	iterator.EXPECT().Next().Return(true).Once()
	iterator.EXPECT().Next().Return(false)
	iterator.EXPECT().Event().Return(nil)
	iterator.EXPECT().Error().Return(nil)
	iterator.EXPECT().Close().Return(nil)

	result := monitor.SyncBlocks(startBlock, endBlock)

	assert.False(t, result)
}

func TestMonitorSyncBlocks_SingleRemovedEvent(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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

	mintControllerAddress := ethcommon.BytesToAddress(mintControllerMap[1])

	startBlock := uint64(1)
	endBlock := uint64(100)

	filter := &bind.FilterOpts{
		Start:   startBlock,
		End:     &endBlock,
		Context: context.Background(),
	}

	iterator := clientMocks.NewMockMailboxDispatchIterator(t)

	event, _ := createValidEvent(t)
	event.Raw.Removed = true

	mailbox.EXPECT().FilterDispatch(filter, []ethcommon.Address{mintControllerAddress}, []uint32{}, [][32]byte{}).Return(iterator, nil)

	iterator.EXPECT().Next().Return(true).Once()
	iterator.EXPECT().Next().Return(false)
	iterator.EXPECT().Event().Return(event)
	iterator.EXPECT().Error().Return(nil)
	iterator.EXPECT().Close().Return(nil)

	result := monitor.SyncBlocks(startBlock, endBlock)

	assert.True(t, result)
}

func TestMonitorSyncBlocks_FilterError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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
	iterator.EXPECT().Error().Return(assert.AnError)
	iterator.EXPECT().Close().Return(nil)

	result := monitor.SyncBlocks(startBlock, endBlock)

	assert.False(t, result)
}

func TestMonitorSyncBlocks_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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

	mailbox.EXPECT().FilterDispatch(filter, []ethcommon.Address{mintControllerAddress}, []uint32{}, [][32]byte{}).Return(nil, assert.AnError)

	result := monitor.SyncBlocks(startBlock, endBlock)

	assert.False(t, result)
}

func TestMonitorSyncNewBlocks(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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

func TestMonitorSyncNewBlocks_MaxQueryBlocks(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	mintControllerMap := map[uint32][]byte{
		1: ethcommon.FromHex("0x01"),
	}

	startBlock := uint64(1)
	endBlock := uint64(100 + eth.MaxQueryBlocks)

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

	iterator := clientMocks.NewMockMailboxDispatchIterator(t)

	mintControllerAddress := ethcommon.BytesToAddress(mintControllerMap[1])

	endBlockOne := startBlock + eth.MaxQueryBlocks

	filterOne := &bind.FilterOpts{
		Start:   startBlock,
		End:     &endBlockOne,
		Context: context.Background(),
	}
	mailbox.EXPECT().FilterDispatch(filterOne, []ethcommon.Address{mintControllerAddress}, []uint32{}, [][32]byte{}).Return(iterator, nil).Once()

	filterTwo := &bind.FilterOpts{
		Start:   endBlockOne,
		End:     &endBlock,
		Context: context.Background(),
	}
	mailbox.EXPECT().FilterDispatch(filterTwo, []ethcommon.Address{mintControllerAddress}, []uint32{}, [][32]byte{}).Return(iterator, nil).Once()

	iterator.EXPECT().Next().Return(false)
	iterator.EXPECT().Error().Return(nil)
	iterator.EXPECT().Close().Return(nil)

	result := monitor.SyncNewBlocks()

	assert.True(t, result)
}

func TestConfirmDispatchTxs(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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

func TestConfirmDispatchTxs_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	mockDB.EXPECT().GetPendingTransactionsTo(mock.Anything, mock.Anything).Return(nil, assert.AnError)

	mailbox.EXPECT().Address().Return(ethcommon.Address{})

	result := monitor.ConfirmDispatchTxs()

	assert.False(t, result)
}

func TestCreateMessagesForTxs_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

	mailbox := clientMocks.NewMockMailboxContract(t)
	monitor := &EthMessageMonitorRunnable{
		db:      mockDB,
		client:  mockClient,
		logger:  logger,
		mailbox: mailbox,
	}

	mockDB.EXPECT().GetConfirmedTransactionsTo(mock.Anything, mock.Anything).Return(nil, assert.AnError)

	mailbox.EXPECT().Address().Return(ethcommon.Address{})

	result := monitor.CreateMessagesForTxs()

	assert.False(t, result)
}

func TestCreateMessagesForTxs(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "monitor")

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
	logger := log.New().WithField("test", "monitor")
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
	logger := log.New().WithField("test", "monitor")

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
