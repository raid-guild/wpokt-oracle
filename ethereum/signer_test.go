package ethereum

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"

	cosmos "github.com/dan13ram/wpokt-oracle/cosmos/client"
	cosmosMocks "github.com/dan13ram/wpokt-oracle/cosmos/client/mocks"
	cosmosUtil "github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/db/mocks"
	clientMocks "github.com/dan13ram/wpokt-oracle/ethereum/client/mocks"
	"github.com/dan13ram/wpokt-oracle/ethereum/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSignerHeight(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")

	monitor := &EthMessageSignerRunnable{
		db:                         mockDB,
		client:                     mockClient,
		logger:                     logger,
		currentEthereumBlockHeight: 100,
	}

	height := monitor.Height()

	assert.Equal(t, uint64(100), height)
}

func TestSignerUpdateCurrentBlockHeight_EthError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &EthMessageSignerRunnable{
		db:              mockDB,
		client:          mockEthClient,
		cosmosClient:    mockCosmosClient,
		logger:          logger,
		timeout:         10 * time.Second,
		numSigners:      1,
		signerThreshold: 1,
	}

	mockEthClient.EXPECT().GetBlockHeight().Return(uint64(100), assert.AnError)
	mockCosmosClient.EXPECT().GetLatestBlockHeight().Return(int64(200), nil)

	signer.UpdateCurrentBlockHeight()

	assert.Equal(t, uint64(0), signer.currentEthereumBlockHeight)
	assert.Equal(t, uint64(200), signer.currentCosmosBlockHeight)
}

func TestSignerUpdateCurrentBlockHeight_CosmosError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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
	mockCosmosClient.EXPECT().GetLatestBlockHeight().Return(int64(200), assert.AnError)

	signer.UpdateCurrentBlockHeight()

	assert.Equal(t, uint64(100), signer.currentEthereumBlockHeight)
	assert.Equal(t, uint64(0), signer.currentCosmosBlockHeight)
}

func TestSignerUpdateCurrentBlockHeight(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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

func TestSignMessage_LockError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

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

	mockDB.EXPECT().LockWriteMessage(message).Return("lock-id", assert.AnError)

	success := signer.SignMessage(message)
	assert.False(t, success)
}

func TestSignMessage_SignError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

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

	utilSignMessage = func(
		msg *models.Message,
		domain util.DomainData,
		privateKey *ecdsa.PrivateKey,
	) error {
		assert.Equal(t, message, msg)
		assert.NotNil(t, domain)
		assert.NotNil(t, privateKey)
		return assert.AnError
	}

	success := signer.SignMessage(message)
	assert.False(t, success)
}

func TestSignMessage_UpdateError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

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
	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(assert.AnError)

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
	assert.False(t, success)
}

func TestSignMessage(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

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

func TestValidateCosmosMessage_GetTxError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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
	mockCosmosClient.EXPECT().GetTx(message.OriginTransactionHash).Return(txResponse, assert.AnError)

	confirmed, err := signer.ValidateCosmosMessage(message)

	assert.False(t, confirmed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error getting tx")
}
func TestValidateCosmosMessage_ValidationError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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
		}, assert.AnError
	}

	confirmed, err := signer.ValidateCosmosMessage(message)

	assert.False(t, confirmed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error validating tx response")
}

func TestValidateCosmosMessage_NeedsRefund(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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
			NeedsRefund:   true,
			Amount:        sdk.NewInt64Coin("uatom", 100),
			SenderAddress: senderAddress.Bytes(),
			TxStatus:      models.TransactionStatusConfirmed,
			Memo: models.MintMemo{
				Address: recipientAddress.Hex(),
			},
		}, nil
	}

	confirmed, err := signer.ValidateCosmosMessage(message)

	assert.False(t, confirmed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tx needs refund")
}

func TestValidateCosmosMessage_AmountMismatch(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))
	message := &models.Message{
		ID: &primitive.ObjectID{},
		Content: models.MessageContent{
			MessageBody: models.MessageBody{
				Amount:           "1000",
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

	assert.False(t, confirmed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount mismatch")
}
func TestValidateCosmosMessage_SenderMismatch(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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
			SenderAddress: []byte("cosmos2"),
			TxStatus:      models.TransactionStatusConfirmed,
			Memo: models.MintMemo{
				Address: recipientAddress.Hex(),
			},
		}, nil
	}

	confirmed, err := signer.ValidateCosmosMessage(message)

	assert.False(t, confirmed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sender mismatch")
}
func TestValidateCosmosMessage_RecipientMismatch(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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
				Address: "0x010204",
			},
		}, nil
	}

	confirmed, err := signer.ValidateCosmosMessage(message)

	assert.False(t, confirmed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recipient mismatch")
}
func TestValidateCosmosMessage_TxPending(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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
			TxStatus:      models.TransactionStatusPending,
			Memo: models.MintMemo{
				Address: recipientAddress.Hex(),
			},
		}, nil
	}

	confirmed, err := signer.ValidateCosmosMessage(message)

	assert.False(t, confirmed)
	assert.NoError(t, err)
}
func TestValidateCosmosMessage_TxInvalid(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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
			TxStatus:      models.TransactionStatusInvalid,
			Memo: models.MintMemo{
				Address: recipientAddress.Hex(),
			},
		}, nil
	}

	confirmed, err := signer.ValidateCosmosMessage(message)

	assert.False(t, confirmed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tx is invalid")
}
func TestValidateCosmosMessage(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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

func TestValidateCosmosTxAndSignMessage_ValidationFailed(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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
			TxStatus:      models.TransactionStatusFailed,
			Memo: models.MintMemo{
				Address: recipientAddress.Hex(),
			},
		}, nil
	}
	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	success := signer.ValidateCosmosTxAndSignMessage(message)
	assert.False(t, success)
}
func TestValidateCosmosTxAndSignMessage_TxPending(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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
			TxStatus:      models.TransactionStatusPending,
			Memo: models.MintMemo{
				Address: recipientAddress.Hex(),
			},
		}, nil
	}

	success := signer.ValidateCosmosTxAndSignMessage(message)
	assert.False(t, success)
}

func TestValidateCosmosTxAndSignMessage(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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

func TestValidateAndFindDispatchIDEvents_MessageIDError(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))
	mailbox := clientMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: mockClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	signer := &EthMessageSignerRunnable{
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		MessageID:             "messageID",
		OriginTransactionHash: "txHash",
		Content: models.MessageContent{
			OriginDomain: 1,
			MessageBody: models.MessageBody{
				SenderAddress:    senderAddress.Hex(),
				RecipientAddress: recipientAddress.Hex(),
				Amount:           "100",
			},
		},
	}

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error getting message ID bytes")
}

func TestValidateAndFindDispatchIDEvents_EthClientError(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))
	mailbox := clientMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	signer := &EthMessageSignerRunnable{
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	txHash := "0x01"
	messageID := ethcommon.BytesToHash([]byte("message1"))

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

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ethereum client not found")
}

func TestValidateAndFindDispatchIDEvents_MailBoxError(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))

	ethClientMap := map[uint32]eth.EthereumClient{1: mockClient}
	mailboxMap := map[uint32]eth.MailboxContract{}

	signer := &EthMessageSignerRunnable{
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	txHash := "0x01"
	messageID := ethcommon.BytesToHash([]byte("message1"))

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

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "mailbox not found")
}

func TestValidateAndFindDispatchIDEvents_ReceiptError(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))
	mailbox := clientMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: mockClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	signer := &EthMessageSignerRunnable{
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	txHash := "0x01"
	messageID := ethcommon.BytesToHash([]byte("message1"))

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

	mockClient.EXPECT().GetTransactionReceipt(txHash).Return(nil, assert.AnError)
	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error getting transaction receipt")
}

func TestValidateAndFindDispatchIDEvents_FailedTxError(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))
	mailbox := clientMocks.NewMockMailboxContract(t)

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
		Status:      types.ReceiptStatusFailed,
		Logs:        []*types.Log{},
	}

	txHash := "0x01"
	messageID := ethcommon.BytesToHash([]byte("message1"))

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
	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.TransactionStatusFailed, result.TxStatus)
}

func TestValidateAndFindDispatchIDEvents_NoEventsError(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
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

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(nil, assert.AnError)

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
	assert.Equal(t, models.TransactionStatusInvalid, result.TxStatus)
	assert.Equal(t, uint64(1), result.Confirmations)
}

func TestValidateAndFindDispatchIDEvents_BlockHeightError(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
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
	mockClient.EXPECT().GetBlockHeight().Return(uint64(101), assert.AnError)

	result, err := signer.ValidateAndFindDispatchIDEvent(message)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error getting current block height")
}

func TestValidateAndFindDispatchIDEvents(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
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

func TestValidateEthereumTxAndSignMessage_ValidateError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))
	mailbox := clientMocks.NewMockMailboxContract(t)

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

	txHash := "0x01"
	messageID := ethcommon.BytesToHash([]byte("message1"))

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

	mockClient.EXPECT().GetTransactionReceipt(txHash).Return(nil, assert.AnError)

	success := signer.ValidateEthereumTxAndSignMessage(message)
	assert.False(t, success)
}
func TestValidateEthereumTxAndSignMessage_PendingTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
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
	mockClient.EXPECT().Confirmations().Return(uint64(1000))

	success := signer.ValidateEthereumTxAndSignMessage(message)
	assert.False(t, success)
}

func TestValidateEthereumTxAndSignMessage_NotConfirmedError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))
	mailbox := clientMocks.NewMockMailboxContract(t)
	mailboxAddress := ethcommon.BytesToAddress([]byte("mailbox1"))

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
		Status:      types.ReceiptStatusFailed,
		Logs: []*types.Log{
			{
				Address: mailboxAddress,
			},
		},
	}

	txHash := "0x01"
	messageID := ethcommon.BytesToHash([]byte("message1"))

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

	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	success := signer.ValidateEthereumTxAndSignMessage(message)
	assert.False(t, success)
}

func TestValidateEthereumTxAndSignMessage(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
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

func TestSignMessages_EthereumTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	logger := log.New().WithField("test", "signer")
	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))
	mailbox := clientMocks.NewMockMailboxContract(t)
	mailboxAddress := ethcommon.BytesToAddress([]byte("mailbox1"))
	mailbox.EXPECT().Address().Return(mailboxAddress)

	ethClientMap := map[uint32]eth.EthereumClient{1: mockClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}
	cosmosClient := cosmosMocks.NewMockCosmosClient(t)

	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	signer := &EthMessageSignerRunnable{
		cosmosClient: cosmosClient,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		db:           mockDB,
		privateKey:   privateKey,
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

	cosmosClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(500)})

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

	mockDB.EXPECT().GetPendingMessages(mock.Anything, mock.Anything).Return([]models.Message{*message}, nil)

	success := signer.SignMessages()
	assert.True(t, success)
}

func TestSignMessages_PrivKeyError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &EthMessageSignerRunnable{
		db:                       mockDB,
		cosmosClient:             mockCosmosClient,
		logger:                   logger,
		signerThreshold:          1,
		privateKey:               nil,
		currentCosmosBlockHeight: 100,
	}

	success := signer.SignMessages()
	assert.False(t, success)
}

func TestSignMessages_ClientError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	signer := &EthMessageSignerRunnable{
		db:                       mockDB,
		cosmosClient:             mockCosmosClient,
		logger:                   logger,
		signerThreshold:          1,
		privateKey:               privateKey,
		currentCosmosBlockHeight: 100,
	}

	mockDB.EXPECT().GetPendingMessages(mock.Anything, mock.Anything).Return(nil, assert.AnError)

	success := signer.SignMessages()
	assert.False(t, success)
}

func TestSignMessages_CosmosTx(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	senderAddress := ethcommon.BytesToAddress([]byte("cosmos1"))
	recipientAddress := ethcommon.BytesToAddress([]byte("eth1"))

	message := &models.Message{
		ID: &primitive.ObjectID{},
		Content: models.MessageContent{
			OriginDomain: 5,
			MessageBody: models.MessageBody{
				SenderAddress:    senderAddress.Hex(),
				RecipientAddress: recipientAddress.Hex(),
				Amount:           "100",
			},
		},
	}

	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	signer := &EthMessageSignerRunnable{
		db:                       mockDB,
		cosmosClient:             mockCosmosClient,
		logger:                   logger,
		signerThreshold:          1,
		privateKey:               privateKey,
		currentCosmosBlockHeight: 100,
	}

	mockCosmosClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(5)})
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

	mockDB.EXPECT().GetPendingMessages(mock.Anything, mock.Anything).Return([]models.Message{*message}, nil)

	success := signer.SignMessages()
	assert.True(t, success)
}

func TestUpdateValidatorCountAndSignerThreshold(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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

	warpISM.EXPECT().ValidatorCount(mock.Anything).Return(big.NewInt(100), nil)
	warpISM.EXPECT().SignerThreshold(mock.Anything).Return(big.NewInt(50), nil)

	signer.UpdateValidatorCountAndSignerThreshold()

	assert.Equal(t, signer.numSigners, int64(100))
	assert.Equal(t, signer.signerThreshold, int64(50))
}

func TestUpdateValidatorCountAndSignerThreshold_ValidatorCountError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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

	warpISM.EXPECT().ValidatorCount(mock.Anything).Return(big.NewInt(100), assert.AnError)

	signer.UpdateValidatorCountAndSignerThreshold()

	assert.Equal(t, signer.numSigners, int64(0))
	assert.Equal(t, signer.signerThreshold, int64(1))
}

func TestUpdateValidatorCountAndSignerThreshold_SignerThresholdError(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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

	warpISM.EXPECT().ValidatorCount(mock.Anything).Return(big.NewInt(100), nil)
	warpISM.EXPECT().SignerThreshold(mock.Anything).Return(big.NewInt(50), assert.AnError)

	signer.UpdateValidatorCountAndSignerThreshold()

	assert.Equal(t, signer.numSigners, int64(100))
	assert.Equal(t, signer.signerThreshold, int64(1))
}

func TestUpdateMaxMintLimit(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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

func TestUpdateMaxMintLimit_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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

	mintController.EXPECT().MaxMintLimit(mock.Anything).Return(big.NewInt(100), assert.AnError)

	signer.UpdateMaxMintLimit()

	var nilAmount *big.Int = nil

	assert.Equal(t, nilAmount, signer.maximumAmount)
}

func TestUpdateDomainData_Error(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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

	warpISM.EXPECT().Eip712Domain(mock.Anything).Return(util.DomainData{Version: "6"}, assert.AnError)

	signer.UpdateDomainData()

	assert.Equal(t, signer.domain, util.DomainData{})
}

func TestUpdateDomainData(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockEthClient := clientMocks.NewMockEthereumClient(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

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
	logger := log.New().WithField("test", "signer")

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
	mockDB.EXPECT().GetPendingMessages(mock.Anything, mock.Anything).Return([]models.Message{}, nil)

	signer.Run()
}

func TestNewMessageSigner(t *testing.T) {
	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	mockMailbox := clientMocks.NewMockMailboxContract(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	mockWarpISM := clientMocks.NewMockWarpISMContract(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)

	mnemonic := "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	config := models.EthereumNetworkConfig{
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
	}

	cosmosNetwork := models.CosmosNetworkConfig{
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

	ethNetworks := []models.EthereumNetworkConfig{
		config,
	}

	ethNewClient = func(models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		return mockClient, nil
	}

	ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
		return mockMailbox, nil
	}

	ethNewMintControllerContract = func(ethcommon.Address, bind.ContractBackend) (eth.MintControllerContract, error) {
		return mockMintController, nil
	}

	ethNewWarpISMContract = func(ethcommon.Address, bind.ContractBackend) (eth.WarpISMContract, error) {
		return mockWarpISM, nil
	}

	cosmosNewClient = func(models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
		return mockCosmosClient, nil
	}

	dbNewDB = func() db.DB {
		return mockDB
	}

	defer func() {
		ethNewClient = eth.NewClient
		ethNewMailboxContract = eth.NewMailboxContract
		ethNewMintControllerContract = eth.NewMintControllerContract
		ethNewWarpISMContract = eth.NewWarpISMContract
		cosmosNewClient = cosmos.NewClient
		dbNewDB = db.NewDB
	}()

	mockClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)

	mockCosmosClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	mockClient.EXPECT().GetClient().Return(nil)

	mockClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(1)})

	mockWarpISM.EXPECT().ValidatorCount(mock.Anything).Return(big.NewInt(3), nil)
	mockWarpISM.EXPECT().SignerThreshold(mock.Anything).Return(big.NewInt(2), nil)
	mockWarpISM.EXPECT().Eip712Domain(mock.Anything).Return(util.DomainData{ChainId: big.NewInt(1), VerifyingContract: ethcommon.HexToAddress(config.WarpISMAddress)}, nil)
	mockMintController.EXPECT().MaxMintLimit(mock.Anything).Return(big.NewInt(100), nil)

	runnable := NewMessageSigner(mnemonic, config, cosmosNetwork, ethNetworks)

	assert.NotNil(t, runnable)

	monitor, ok := runnable.(*EthMessageSignerRunnable)
	assert.True(t, ok)
	assert.Equal(t, mockDB, monitor.db)
	assert.Equal(t, mockClient, monitor.client)
	assert.Equal(t, mockMintController, monitor.mintController)
	assert.Equal(t, mockCosmosClient, monitor.cosmosClient)
	assert.Equal(t, uint64(100), monitor.currentEthereumBlockHeight)
	assert.Equal(t, uint64(100), monitor.currentCosmosBlockHeight)

}

func TestNewMessageSignerFailures(t *testing.T) {

	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	mockDB := mocks.NewMockDB(t)
	mockClient := clientMocks.NewMockEthereumClient(t)
	mockMailbox := clientMocks.NewMockMailboxContract(t)
	mockMintController := clientMocks.NewMockMintControllerContract(t)
	mockWarpISM := clientMocks.NewMockWarpISMContract(t)
	mockCosmosClient := cosmosMocks.NewMockCosmosClient(t)

	mnemonic := "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	config := models.EthereumNetworkConfig{
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
	}

	cosmosNetwork := models.CosmosNetworkConfig{
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

	ethNetworks := []models.EthereumNetworkConfig{
		config,
	}

	ethNewClient = func(config models.EthereumNetworkConfig) (eth.EthereumClient, error) {
		return mockClient, nil
	}

	ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
		return mockMailbox, nil
	}

	ethNewMintControllerContract = func(ethcommon.Address, bind.ContractBackend) (eth.MintControllerContract, error) {
		return mockMintController, nil
	}

	ethNewWarpISMContract = func(ethcommon.Address, bind.ContractBackend) (eth.WarpISMContract, error) {
		return mockWarpISM, nil
	}

	cosmosNewClient = func(models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
		return mockCosmosClient, nil
	}

	dbNewDB = func() db.DB {
		return mockDB
	}

	defer func() {
		ethNewClient = eth.NewClient
		ethNewMailboxContract = eth.NewMailboxContract
		ethNewMintControllerContract = eth.NewMintControllerContract
		ethNewWarpISMContract = eth.NewWarpISMContract
		cosmosNewClient = cosmos.NewClient
		dbNewDB = db.NewDB
	}()

	mockClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)

	mockCosmosClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	mockClient.EXPECT().GetClient().Return(nil)

	mockClient.EXPECT().Chain().Return(models.Chain{ChainDomain: uint32(1)})

	t.Run("Disabled", func(t *testing.T) {
		config.MessageSigner.Enabled = false

		assert.Panics(t, func() {
			NewMessageSigner(mnemonic, config, cosmosNetwork, ethNetworks)
		})

		config.MessageSigner.Enabled = true
	})

	t.Run("ClientError", func(t *testing.T) {

		ethNewClient = func(models.EthereumNetworkConfig) (eth.EthereumClient, error) {
			return nil, assert.AnError
		}

		assert.Panics(t, func() {
			NewMessageSigner(mnemonic, config, cosmosNetwork, ethNetworks)
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
			NewMessageSigner(mnemonic, config, cosmosNetwork, ethNetworks)
		})

		ethNewMintControllerContract = func(ethcommon.Address, bind.ContractBackend) (eth.MintControllerContract, error) {
			return mockMintController, nil
		}

	})

	t.Run("WarpISMError", func(t *testing.T) {

		ethNewWarpISMContract = func(ethcommon.Address, bind.ContractBackend) (eth.WarpISMContract, error) {
			return nil, assert.AnError
		}

		assert.Panics(t, func() {
			NewMessageSigner(mnemonic, config, cosmosNetwork, ethNetworks)
		})

		ethNewWarpISMContract = func(ethcommon.Address, bind.ContractBackend) (eth.WarpISMContract, error) {
			return mockWarpISM, nil
		}

	})

	t.Run("MnemonicError", func(t *testing.T) {

		mnemonic = "invalid"

		assert.Panics(t, func() {
			NewMessageSigner(mnemonic, config, cosmosNetwork, ethNetworks)
		})

		mnemonic = "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve"

	})

	t.Run("CosmosClientError", func(t *testing.T) {

		cosmosNewClient = func(models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
			return nil, assert.AnError
		}

		assert.Panics(t, func() {
			NewMessageSigner(mnemonic, config, cosmosNetwork, ethNetworks)
		})

		cosmosNewClient = func(models.CosmosNetworkConfig) (cosmos.CosmosClient, error) {
			return mockCosmosClient, nil
		}

	})

	t.Run("SecondEthClientError", func(t *testing.T) {
		newEthNetworks := append(ethNetworks, models.EthereumNetworkConfig{
			ChainID: 2,
		})

		assert.Equal(t, 2, len(newEthNetworks))
		assert.Equal(t, 1, len(ethNetworks))

		ethNewClient = func(config models.EthereumNetworkConfig) (eth.EthereumClient, error) {
			if config.ChainID == 1 {
				return mockClient, nil
			}
			return nil, assert.AnError
		}

		assert.Panics(t, func() {
			NewMessageSigner(mnemonic, config, cosmosNetwork, newEthNetworks)
		})

		ethNewClient = func(config models.EthereumNetworkConfig) (eth.EthereumClient, error) {
			return mockClient, nil
		}

	})

	t.Run("MailboxError", func(t *testing.T) {

		ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
			return nil, assert.AnError
		}

		assert.Panics(t, func() {
			NewMessageSigner(mnemonic, config, cosmosNetwork, ethNetworks)
		})

		ethNewMailboxContract = func(ethcommon.Address, bind.ContractBackend) (eth.MailboxContract, error) {
			return mockMailbox, nil
		}

	})

	t.Run("InvalidSigners", func(t *testing.T) {
		mockWarpISM.EXPECT().ValidatorCount(mock.Anything).Return(big.NewInt(100), nil).Once()
		mockWarpISM.EXPECT().SignerThreshold(mock.Anything).Return(big.NewInt(20), nil).Once()
		assert.Panics(t, func() {
			NewMessageSigner(mnemonic, config, cosmosNetwork, ethNetworks)
		})
	})

	t.Run("InvalidThreshold", func(t *testing.T) {
		mockWarpISM.EXPECT().ValidatorCount(mock.Anything).Return(big.NewInt(3), nil).Once()
		mockWarpISM.EXPECT().SignerThreshold(mock.Anything).Return(big.NewInt(20), nil).Once()
		assert.Panics(t, func() {
			NewMessageSigner(mnemonic, config, cosmosNetwork, ethNetworks)
		})
	})

	t.Run("DomainDataErrorChainID", func(t *testing.T) {
		mockWarpISM.EXPECT().ValidatorCount(mock.Anything).Return(big.NewInt(3), nil).Once()
		mockWarpISM.EXPECT().SignerThreshold(mock.Anything).Return(big.NewInt(2), nil).Once()
		mockWarpISM.EXPECT().Eip712Domain(mock.Anything).Return(util.DomainData{ChainId: big.NewInt(2), VerifyingContract: ethcommon.HexToAddress(config.WarpISMAddress)}, nil).Once()
		assert.Panics(t, func() {
			NewMessageSigner(mnemonic, config, cosmosNetwork, ethNetworks)
		})
	})

	t.Run("DomainDataErrorContract", func(t *testing.T) {
		mockWarpISM.EXPECT().ValidatorCount(mock.Anything).Return(big.NewInt(3), nil).Once()
		mockWarpISM.EXPECT().SignerThreshold(mock.Anything).Return(big.NewInt(2), nil).Once()
		mockWarpISM.EXPECT().Eip712Domain(mock.Anything).Return(util.DomainData{ChainId: big.NewInt(1), VerifyingContract: ethcommon.BytesToAddress([]byte("invalid"))}, nil).Once()
		assert.Panics(t, func() {
			NewMessageSigner(mnemonic, config, cosmosNetwork, ethNetworks)
		})
	})

}
