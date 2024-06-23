package ethereum

import (
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"

	cosmosMocks "github.com/dan13ram/wpokt-oracle/cosmos/client/mocks"
	cosmosUtil "github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/db/mocks"
	clientMocks "github.com/dan13ram/wpokt-oracle/ethereum/client/mocks"
	"github.com/dan13ram/wpokt-oracle/ethereum/util"
	"github.com/dan13ram/wpokt-oracle/models"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"

	// "github.com/dan13ram/wpokt-oracle/common"
	// "github.com/ethereum/go-ethereum/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestUpdateCurrentBlockHeight(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "signer")

	signer := &EthMessageSignerRunnable{
		db:              mockDB,
		client:          mockEthClient,
		cosmosClient:    mockCosmosClient,
		logger:          logger,
		timeout:         10 * time.Second,
		numSigners:      1,
		signerThreshold: 1,
	}

	mockEthClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	mockCosmosClient.EXPECT().GetLatestBlockHeight().Return(int64(200), nil)

	signer.UpdateCurrentBlockHeight()

	assert.Equal(t, uint64(100), signer.currentEthereumBlockHeight)
	assert.Equal(t, uint64(200), signer.currentCosmosBlockHeight)
}

func TestSignMessage(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := logrus.New().WithField("test", "signer")

	message := &models.Message{
		ID: &primitive.ObjectID{},
		Content: models.MessageContent{
			MessageBody: models.MessageBody{
				Amount: "100",
			},
		},
	}

	signer := &EthMessageSignerRunnable{
		db:              mockDB,
		logger:          logger,
		signerThreshold: 1,
		privateKey:      &ecdsa.PrivateKey{},
	}

	mockDB.EXPECT().LockWriteMessage(message).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	utilSignMessage = func(
		msg *models.Message,
		domain util.DomainData,
		privateKey *ecdsa.PrivateKey,
	) error {
		assert.Equal(t, message, msg)
		assert.NotNil(t, domain)
		assert.NotNil(t, privateKey)
		return nil
	}

	success := signer.SignMessage(message)
	assert.True(t, success)
}

func TestValidateCosmosMessage(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "signer")

	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))
	message := &models.Message{
		ID: &primitive.ObjectID{},
		Content: models.MessageContent{
			MessageBody: models.MessageBody{
				Amount:           "100",
				SenderAddress:    senderAddress.Hex(),
				RecipientAddress: recipientAddress.Hex(),
			},
		},
	}

	signer := &EthMessageSignerRunnable{
		db:                       mockDB,
		cosmosClient:             mockCosmosClient,
		logger:                   logger,
		signerThreshold:          1,
		privateKey:               &ecdsa.PrivateKey{},
		currentCosmosBlockHeight: 100,
	}

	txResponse := &sdk.TxResponse{}
	mockCosmosClient.EXPECT().GetTx(message.OriginTransactionHash).Return(txResponse, nil)

	utilValidateTxToCosmosMultisig = func(
		txResponse *sdk.TxResponse,
		config models.CosmosNetworkConfig,
		supportedChainIDsEthereum map[uint32]bool,
		currentCosmosBlockHeight uint64,
	) (*cosmosUtil.ValidateTxResult, error) {
		assert.NotNil(t, txResponse)
		assert.NotNil(t, config)
		assert.NotNil(t, supportedChainIDsEthereum)
		assert.Equal(t, uint64(100), currentCosmosBlockHeight)
		return &cosmosUtil.ValidateTxResult{
			Amount:        sdk.NewInt64Coin("uatom", 100),
			SenderAddress: senderAddress.Bytes(),
			TxStatus:      models.TransactionStatusConfirmed,
			Memo: models.MintMemo{
				Address: recipientAddress.Hex(),
			},
		}, nil
	}

	confirmed, err := signer.ValidateCosmosMessage(message)

	assert.True(t, confirmed)
	assert.NoError(t, err)
}

func TestValidateCosmosTxAndSignMessage(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "signer")

	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))

	message := &models.Message{
		ID: &primitive.ObjectID{},
		Content: models.MessageContent{
			MessageBody: models.MessageBody{
				SenderAddress:    senderAddress.Hex(),
				RecipientAddress: recipientAddress.Hex(),
				Amount:           "100",
			},
		},
	}

	signer := &EthMessageSignerRunnable{
		db:                       mockDB,
		cosmosClient:             mockCosmosClient,
		logger:                   logger,
		signerThreshold:          1,
		privateKey:               &ecdsa.PrivateKey{},
		currentCosmosBlockHeight: 100,
	}

	txResponse := &sdk.TxResponse{
		Height: 50,
	}
	mockCosmosClient.EXPECT().GetTx(message.OriginTransactionHash).Return(txResponse, nil)

	utilValidateTxToCosmosMultisig = func(
		txResponse *sdk.TxResponse,
		config models.CosmosNetworkConfig,
		supportedChainIDsEthereum map[uint32]bool,
		currentCosmosBlockHeight uint64,
	) (*cosmosUtil.ValidateTxResult, error) {
		assert.NotNil(t, txResponse)
		assert.NotNil(t, config)
		assert.NotNil(t, supportedChainIDsEthereum)
		assert.Equal(t, uint64(100), currentCosmosBlockHeight)
		return &cosmosUtil.ValidateTxResult{
			Amount:        sdk.NewInt64Coin("uatom", 100),
			SenderAddress: senderAddress.Bytes(),
			TxStatus:      models.TransactionStatusConfirmed,
			Memo: models.MintMemo{
				Address: recipientAddress.Hex(),
			},
		}, nil
	}
	utilSignMessage = func(
		msg *models.Message,
		domain util.DomainData,
		privateKey *ecdsa.PrivateKey,
	) error {
		assert.Equal(t, message, msg)
		assert.NotNil(t, domain)
		assert.NotNil(t, privateKey)
		return nil
	}

	mockDB.EXPECT().LockWriteMessage(message).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	success := signer.ValidateCosmosTxAndSignMessage(message)
	assert.True(t, success)
}

func TestValidateAndFindDispatchIDEvents(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "signer")
	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))
	mailbox := clientMocks.NewMockMailboxContract(t)
	mailboxAddress := ethcommon.BytesToAddress([]byte("mailbox1"))
	mailbox.EXPECT().Address().Return(mailboxAddress)

	ethClientMap := map[uint32]eth.EthereumClient{1: mockClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	signer := &EthMessageSignerRunnable{
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	receipt := &types.Receipt{
		BlockNumber: big.NewInt(100),
		Status:      types.ReceiptStatusSuccessful,
		Logs: []*types.Log{
			{
				Address: mailboxAddress,
			},
		},
	}

	txHash := "0x01"
	messageID := ethcommon.BytesToHash([]byte("message1"))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: messageID}, nil)

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		MessageID:             messageID.Hex(),
		OriginTransactionHash: txHash,
		Content: models.MessageContent{
			OriginDomain: 1,
			MessageBody: models.MessageBody{
				SenderAddress:    senderAddress.Hex(),
				RecipientAddress: recipientAddress.Hex(),
				Amount:           "100",
			},
		},
	}

	mockClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil)
	mockClient.EXPECT().GetBlockHeight().Return(uint64(101), nil)
	mockClient.EXPECT().Confirmations().Return(uint64(1))

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.TransactionStatusConfirmed, result.TxStatus)
	assert.Equal(t, uint64(1), result.Confirmations)
}

func TestValidateEthereumTxAndSignMessage(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := logrus.New().WithField("test", "signer")
	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))
	mailbox := clientMocks.NewMockMailboxContract(t)
	mailboxAddress := ethcommon.BytesToAddress([]byte("mailbox1"))
	mailbox.EXPECT().Address().Return(mailboxAddress)

	ethClientMap := map[uint32]eth.EthereumClient{1: mockClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	signer := &EthMessageSignerRunnable{
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		db:           mockDB,
		privateKey:   &ecdsa.PrivateKey{},
	}

	receipt := &types.Receipt{
		BlockNumber: big.NewInt(100),
		Status:      types.ReceiptStatusSuccessful,
		Logs: []*types.Log{
			{
				Address: mailboxAddress,
			},
		},
	}

	txHash := "0x01"
	messageID := ethcommon.BytesToHash([]byte("message1"))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: messageID}, nil)

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		MessageID:             messageID.Hex(),
		OriginTransactionHash: txHash,
		Content: models.MessageContent{
			OriginDomain: 1,
			MessageBody: models.MessageBody{
				SenderAddress:    senderAddress.Hex(),
				RecipientAddress: recipientAddress.Hex(),
				Amount:           "100",
			},
		},
	}

	mockClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil)
	mockClient.EXPECT().GetBlockHeight().Return(uint64(101), nil)
	mockClient.EXPECT().Confirmations().Return(uint64(1))

	mockDB.EXPECT().LockWriteMessage(message).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	utilSignMessage = func(
		msg *models.Message,
		domain util.DomainData,
		privateKey *ecdsa.PrivateKey,
	) error {
		assert.Equal(t, message, msg)
		assert.NotNil(t, domain)
		assert.NotNil(t, privateKey)
		return nil
	}

	success := signer.ValidateEthereumTxAndSignMessage(message)
	assert.True(t, success)
}

func TestUpdateMaxMintLimit(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "signer")

	mintController := clientMocks.NewMockMintControllerContract(t)

	signer := &EthMessageSignerRunnable{
		db:              mockDB,
		client:          mockEthClient,
		cosmosClient:    mockCosmosClient,
		logger:          logger,
		timeout:         10 * time.Second,
		signerThreshold: 1,
		privateKey:      &ecdsa.PrivateKey{},
		mintController:  mintController,
	}

	mintController.EXPECT().MaxMintLimit(mock.Anything).Return(big.NewInt(100), nil)

	signer.UpdateMaxMintLimit()

	assert.Equal(t, signer.maximumAmount, big.NewInt(100))
}

func TestUpdateDomainData(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "signer")

	warpISM := clientMocks.NewMockWarpISMContract(t)

	signer := &EthMessageSignerRunnable{
		db:              mockDB,
		client:          mockEthClient,
		cosmosClient:    mockCosmosClient,
		logger:          logger,
		timeout:         10 * time.Second,
		signerThreshold: 1,
		privateKey:      &ecdsa.PrivateKey{},
		warpISM:         warpISM,
	}

	warpISM.EXPECT().Eip712Domain(mock.Anything).Return(util.DomainData{Version: "6"}, nil)

	signer.UpdateDomainData()

	assert.Equal(t, signer.domain, util.DomainData{Version: "6"})
}

func TestSignerRun(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := logrus.New().WithField("test", "signer")

	mintController := clientMocks.NewMockMintControllerContract(t)
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	signer := &EthMessageSignerRunnable{
		db:              mockDB,
		client:          mockEthClient,
		cosmosClient:    mockCosmosClient,
		logger:          logger,
		timeout:         10 * time.Second,
		signerThreshold: 1,
		privateKey:      privateKey,
		mintController:  mintController,
	}

	mockEthClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	mockCosmosClient.EXPECT().GetLatestBlockHeight().Return(int64(200), nil)
	mintController.EXPECT().MaxMintLimit(mock.Anything).Return(big.NewInt(100), nil)
	mockDB.EXPECT().GetPendingMessages(mock.Anything, mock.Anything).Return([]models.Message{}, nil)

	signer.Run()

	assert.Equal(t, signer.maximumAmount, big.NewInt(100))
}
