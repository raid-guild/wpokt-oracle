package cosmos

import (
	"context"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/cosmos/cosmos-sdk/client"

	clientMocks "github.com/dan13ram/wpokt-oracle/cosmos/client/mocks"
	dbMocks "github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/models"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	ethcommon "github.com/ethereum/go-ethereum/common"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	// authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/core/types"

	// "github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	ethMocks "github.com/dan13ram/wpokt-oracle/ethereum/client/mocks"

	log "github.com/sirupsen/logrus"
)

func TestSignerHeight(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:                 mockDB,
		client:             mockClient,
		logger:             logger,
		currentBlockHeight: 100,
	}

	height := signer.Height()

	assert.Equal(t, uint64(100), height)
}

func TestSignerUpdateCurrentHeight(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), nil)

	signer.UpdateCurrentHeight()

	mockClient.AssertExpectations(t)
	assert.Equal(t, uint64(100), signer.currentBlockHeight)
}

func TestSignerUpdateCurrentHeight_Error(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockClient.EXPECT().GetLatestBlockHeight().Return(int64(100), assert.AnError)

	signer.UpdateCurrentHeight()

	mockClient.AssertExpectations(t)
	assert.Equal(t, uint64(0), signer.currentBlockHeight)
}

func TestSignerUpdateMessage(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

	message := &models.Message{ID: &primitive.ObjectID{}}
	update := bson.M{"status": models.MessageStatusSigned}

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateMessage(message.ID, update).Return(nil)

	result := signer.UpdateMessage(message, update)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignerUpdateMessage_Error(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

	message := &models.Message{ID: &primitive.ObjectID{}}
	update := bson.M{"status": models.MessageStatusSigned}

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateMessage(message.ID, update).Return(assert.AnError)

	result := signer.UpdateMessage(message, update)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignMessage(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:         mockDB,
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
		signerKey:  signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: "multisigAddress",
		},
	}

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)
	tx := clientMocks.NewMockTx(t)

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	utilNewSendTx = func(string, []byte, []byte, sdk.Coin, string, sdk.Coin) (string, error) {
		return "txBody", nil
	}

	utilWrapTxBuilder = func(prefix string, txBody string) (client.TxBuilder, client.TxConfig, error) {
		assert.Equal(t, "pokt", prefix)
		assert.Equal(t, "txBody", txBody)
		return txBuilder, txConfig, nil
	}

	utilSignWithPrivKey = func(context.Context, signing.SignerData, client.TxBuilder, cryptotypes.PrivKey, client.TxConfig, uint64) (signingtypes.SignatureV2, []byte, error) {
		return signingtypes.SignatureV2{
			PubKey: signerKey.PubKey(),
			Data: &signingtypes.SingleSignatureData{
				SignMode:  signingtypes.SignMode_SIGN_MODE_DIRECT,
				Signature: []byte("signature"),
			},
		}, nil, nil
	}

	signers := [][]byte{multisigPk.Address().Bytes()}

	assert.True(t, isTxSigner(multisigPk.Address().Bytes(), signers))

	tx.EXPECT().GetSigners().Return(signers, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)
	txBuilder.EXPECT().GetTx().Return(tx)

	var encoder sdk.TxEncoder = func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	}

	txConfig.EXPECT().TxJSONEncoder().Return(encoder)

	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	mockClient.EXPECT().GetAccount("multisigAddress").Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	result := signer.SignMessage(message)

	mockDB.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.True(t, result)
}

func TestValidateAndFindDispatchIdEvent_InvalidMessageID(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             "invalid",
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	result, err := signer.ValidateAndFindDispatchIdEvent(message)

	assert.Nil(t, result)
	assert.Error(t, err)
	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
}

func TestValidateAndFindDispatchIdEvent_EthClientError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	result, err := signer.ValidateAndFindDispatchIdEvent(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestValidateAndFindDispatchIdEvent_MailboxError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	result, err := signer.ValidateAndFindDispatchIdEvent(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestValidateAndFindDispatchIdEvent_ReceiptError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(nil, assert.AnError)

	result, err := signer.ValidateAndFindDispatchIdEvent(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestValidateAndFindDispatchIdEvent_UnsuccessfulReceipt(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusFailed,
		Logs:        []*types.Log{},
	}, nil).Once()

	result, err := signer.ValidateAndFindDispatchIdEvent(message)

	assert.NotNil(t, result)
	assert.Equal(t, models.TransactionStatusFailed, result.TxStatus)
	assert.NoError(t, err)

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(nil, nil).Once()

	result, err = signer.ValidateAndFindDispatchIdEvent(message)

	assert.NotNil(t, result)
	assert.Equal(t, models.TransactionStatusFailed, result.TxStatus)
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
}

func TestValidateAndFindDispatchIdEvent_BlockHeightError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), assert.AnError)
	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	result, err := signer.ValidateAndFindDispatchIdEvent(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestValidateAndFindDispatchIdEvent_NoEvent(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(nil, assert.AnError)

	result, err := signer.ValidateAndFindDispatchIdEvent(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), result.Confirmations)
	assert.Equal(t, models.TransactionStatusInvalid, result.TxStatus)
}

func TestValidateAndFindDispatchIdEvent(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	result, err := signer.ValidateAndFindDispatchIdEvent(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Equal(t, messageIDBytes, result.Event.MessageId)
	assert.Equal(t, uint64(10), result.Confirmations)
	assert.Equal(t, models.TransactionStatusConfirmed, result.TxStatus)
}

func TestValidateEthereumTxAndSignMessage_ValidateError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             "0xinvalid",
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: "multisigAddress",
		},
	}

	result := signer.ValidateEthereumTxAndSignMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateEthereumTxAndSignMessage_Pending(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: "multisigAddress",
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(91), nil)
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	result := signer.ValidateEthereumTxAndSignMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateEthereumTxAndSignMessage_FailedTx(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: "multisigAddress",
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusFailed,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)

	mockDB.EXPECT().UpdateMessage(message.ID, bson.M{"status": models.MessageStatusInvalid}).Return(nil)

	result := signer.ValidateEthereumTxAndSignMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.True(t, result)
}

func TestValidateEthereumTxAndSignMessage_LockError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: "multisigAddress",
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	mockDB.EXPECT().LockWriteMessage(mock.Anything).Return("lock-id", assert.AnError)

	result := signer.ValidateEthereumTxAndSignMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.False(t, result)
}

func TestValidateEthereumTxAndSignMessage(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: "multisigAddress",
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)
	tx := clientMocks.NewMockTx(t)

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	utilNewSendTx = func(string, []byte, []byte, sdk.Coin, string, sdk.Coin) (string, error) {
		return "txBody", nil
	}

	utilWrapTxBuilder = func(prefix string, txBody string) (client.TxBuilder, client.TxConfig, error) {
		assert.Equal(t, "pokt", prefix)
		assert.Equal(t, "txBody", txBody)
		return txBuilder, txConfig, nil
	}

	utilSignWithPrivKey = func(context.Context, signing.SignerData, client.TxBuilder, cryptotypes.PrivKey, client.TxConfig, uint64) (signingtypes.SignatureV2, []byte, error) {
		return signingtypes.SignatureV2{
			PubKey: signerKey.PubKey(),
			Data: &signingtypes.SingleSignatureData{
				SignMode:  signingtypes.SignMode_SIGN_MODE_DIRECT,
				Signature: []byte("signature"),
			},
		}, nil, nil
	}

	signers := [][]byte{multisigPk.Address().Bytes()}

	assert.True(t, isTxSigner(multisigPk.Address().Bytes(), signers))

	tx.EXPECT().GetSigners().Return(signers, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)
	txBuilder.EXPECT().GetTx().Return(tx)

	var encoder sdk.TxEncoder = func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	}

	txConfig.EXPECT().TxJSONEncoder().Return(encoder)

	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	mockClient.EXPECT().GetAccount("multisigAddress").Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	mockDB.EXPECT().LockWriteMessage(mock.Anything).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	result := signer.ValidateEthereumTxAndSignMessage(message)

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignMessages(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	ethClient := ethMocks.NewMockEthereumClient(t)
	mailbox := ethMocks.NewMockMailboxContract(t)

	mailbox.EXPECT().Address().Return(ethcommon.BytesToAddress([]byte("mailbox")))
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	ethClientMap := map[uint32]eth.EthereumClient{1: ethClient}
	mailboxMap := map[uint32]eth.MailboxContract{1: mailbox}

	messageIDBytes := [32]byte{}
	messageID := common.HexFromBytes(messageIDBytes[:])

	recipientAddr := ethcommon.BytesToAddress([]byte("recipient"))

	message := &models.Message{
		ID:                    &primitive.ObjectID{},
		OriginTransactionHash: "hash1",
		Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: recipientAddr.Hex(), Amount: "100"}},
		Signatures:            []models.Signature{},
		MessageID:             messageID,
		Sequence:              new(uint64),
	}

	signer := &CosmosMessageSignerRunnable{
		db:           mockDB,
		client:       mockClient,
		logger:       logger,
		ethClientMap: ethClientMap,
		mailboxMap:   mailboxMap,
		multisigPk:   multisigPk,
		signerKey:    signerKey,
		config: models.CosmosNetworkConfig{
			ChainID:         "chain-id",
			CoinDenom:       "upokt",
			Bech32Prefix:    "pokt",
			MultisigAddress: "multisigAddress",
		},
	}

	ethClient.EXPECT().GetTransactionReceipt("hash1").Return(&types.Receipt{
		BlockNumber: big.NewInt(90),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        []*types.Log{{Address: mailbox.Address()}},
	}, nil)
	ethClient.EXPECT().GetBlockHeight().Return(uint64(100), nil)
	ethClient.EXPECT().Confirmations().Return(uint64(10))

	mailbox.EXPECT().ParseDispatchId(mock.Anything).Return(&autogen.MailboxDispatchId{MessageId: [32]byte{}}, nil)

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)
	tx := clientMocks.NewMockTx(t)

	mockDB.EXPECT().LockWriteSequence().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	utilNewSendTx = func(string, []byte, []byte, sdk.Coin, string, sdk.Coin) (string, error) {
		return "txBody", nil
	}

	utilWrapTxBuilder = func(prefix string, txBody string) (client.TxBuilder, client.TxConfig, error) {
		assert.Equal(t, "pokt", prefix)
		assert.Equal(t, "txBody", txBody)
		return txBuilder, txConfig, nil
	}

	utilSignWithPrivKey = func(context.Context, signing.SignerData, client.TxBuilder, cryptotypes.PrivKey, client.TxConfig, uint64) (signingtypes.SignatureV2, []byte, error) {
		return signingtypes.SignatureV2{
			PubKey: signerKey.PubKey(),
			Data: &signingtypes.SingleSignatureData{
				SignMode:  signingtypes.SignMode_SIGN_MODE_DIRECT,
				Signature: []byte("signature"),
			},
		}, nil, nil
	}

	signers := [][]byte{multisigPk.Address().Bytes()}

	assert.True(t, isTxSigner(multisigPk.Address().Bytes(), signers))

	tx.EXPECT().GetSigners().Return(signers, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)
	txBuilder.EXPECT().GetTx().Return(tx)

	var encoder sdk.TxEncoder = func(tx sdk.Tx) ([]byte, error) {
		return []byte("encoded tx"), nil
	}

	txConfig.EXPECT().TxJSONEncoder().Return(encoder)

	mockDB.EXPECT().UpdateMessage(message.ID, mock.Anything).Return(nil)

	mockClient.EXPECT().GetAccount("multisigAddress").Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	mockDB.EXPECT().LockWriteMessage(mock.Anything).Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)

	mockDB.EXPECT().GetPendingMessages(mock.Anything, mock.Anything).Return([]models.Message{*message}, nil)

	result := signer.SignMessages()

	mockDB.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	mailbox.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignMessages_DBError(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signerKey := secp256k1.GenPrivKey()
	signer := &CosmosMessageSignerRunnable{
		db:        mockDB,
		client:    mockClient,
		logger:    logger,
		signerKey: signerKey,
	}

	mockDB.EXPECT().GetPendingMessages(mock.Anything, mock.Anything).Return(nil, assert.AnError)

	result := signer.SignMessages()

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestSignerUpdateRefund(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

	refund := &models.Refund{ID: &primitive.ObjectID{}}
	update := bson.M{"status": models.RefundStatusSigned}

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateRefund(refund.ID, update).Return(nil)

	result := signer.UpdateRefund(refund, update)

	mockDB.AssertExpectations(t)
	assert.True(t, result)
}

func TestSignerUpdateRefund_Error(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	logger := log.New().WithField("test", "signer")

	refund := &models.Refund{ID: &primitive.ObjectID{}}
	update := bson.M{"status": models.RefundStatusSigned}

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		logger: logger,
	}

	mockDB.EXPECT().UpdateRefund(refund.ID, update).Return(assert.AnError)

	result := signer.UpdateRefund(refund, update)

	mockDB.AssertExpectations(t)
	assert.False(t, result)
}

func TestIsTxSigner(t *testing.T) {
	signer := secp256k1.GenPrivKey()
	signerPk := signer.PubKey()

	signers := [][]byte{signerPk.Bytes()}

	assert.True(t, isTxSigner(signerPk.Bytes(), signers))

	otherSigner := secp256k1.GenPrivKey()

	assert.False(t, isTxSigner(otherSigner.PubKey().Bytes(), signers))
}

/*
	func TestBroadcastMessage(t *testing.T) {
		mockDB := dbMocks.NewMockDB(t)
		mockClient := clientMocks.NewMockCosmosClient(t)
		logger := log.New().WithField("test", "signer")

		message := &models.Message{
			ID:                    &primitive.ObjectID{},
			OriginTransactionHash: "hash1",
			Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
			Signatures:            []models.Signature{},
		}

		signer := &CosmosMessageSignerRunnable{
			db:     mockDB,
			client: mockClient,
			logger: logger,
		}

		txBuilder := clientMocks.NewMockTxBuilder(t)
		txConfig := clientMocks.NewMockTxConfig(t)

		utilWrapTxBuilder = func(bech32Prefix, txBody string) (client.TxBuilder, client.TxConfig, error) {
			return txBuilder, txConfig, nil
		}

		txBuilder.EXPECT().SetSignatures( mock.Anything).Return(nil)
		txBuilder.EXPECT().GetTx().Return(mock.Anything)
		txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
			return []byte("{}"), nil
		})

		mockClient.EXPECT().BroadcastTx", mock.Anything).Return("txHash( nil)
		mockDB.EXPECT().UpdateMessage( message.ID, mock.Anything).Return(nil)

		result := signer.BroadcastMessage(message)

		mockClient.AssertExpectations(t)
		mockDB.AssertExpectations(t)
		txBuilder.AssertExpectations(t)
		txConfig.AssertExpectations(t)
		assert.True(t, result)
	}

	func TestBroadcastMessages(t *testing.T) {
		mockDB := dbMocks.NewMockDB(t)
		mockClient := clientMocks.NewMockCosmosClient(t)
		logger := log.New().WithField("test", "signer")

		message := &models.Message{
			ID:                    &primitive.ObjectID{},
			OriginTransactionHash: "hash1",
			Content:               models.MessageContent{OriginDomain: 1, MessageBody: models.MessageBody{RecipientAddress: "recipient"}},
			Signatures:            []models.Signature{},
		}

		signer := &CosmosMessageSignerRunnable{
			db:     mockDB,
			client: mockClient,
			logger: logger,
		}

		mockDB.EXPECT().GetSignedMessages( mock.Anything).Return([]models.Message{*message}, nil)
		mockDB.EXPECT().UpdateMessage( message.ID, mock.Anything).Return(nil)

		utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
			return &util.ValidateTxResult{
				Confirmations: 2,
				TxStatus:      models.TransactionStatusConfirmed,
			}, nil
		}

		result := signer.BroadcastMessages()

		mockDB.AssertExpectations(t)
		assert.True(t, result)
	}

	func TestSignRefunds(t *testing.T) {
		mockDB := dbMocks.NewMockDB(t)
		mockClient := clientMocks.NewMockCosmosClient(t)
		logger := log.New().WithField("test", "signer")

		refund := &models.Refund{
			ID:                    &primitive.ObjectID{},
			OriginTransactionHash: "hash1",
			Recipient:             "recipient",
			Amount:                "100",
		}

		signer := &CosmosMessageSignerRunnable{
			db:     mockDB,
			client: mockClient,
			logger: logger,
		}

		mockDB.EXPECT().GetPendingRefunds( mock.Anything).Return([]models.Refund{*refund}, nil)
		mockDB.EXPECT().UpdateRefund( refund.ID, mock.Anything).Return(nil)

		utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
			return &util.ValidateTxResult{
				Confirmations: 2,
				TxStatus:      models.TransactionStatusConfirmed,
			}, nil
		}

		result := signer.SignRefunds()

		mockDB.AssertExpectations(t)
		assert.True(t, result)
	}

	func TestBroadcastRefund(t *testing.T) {
		mockDB := dbMocks.NewMockDB(t)
		mockClient := clientMocks.NewMockCosmosClient(t)
		logger := log.New().WithField("test", "signer")

		refund := &models.Refund{
			ID:                    &primitive.ObjectID{},
			OriginTransactionHash: "hash1",
			Recipient:             "recipient",
			Amount:                "100",
		}

		signer := &CosmosMessageSignerRunnable{
			db:     mockDB,
			client: mockClient,
			logger: logger,
		}

		txBuilder := clientMocks.NewMockTxBuilder(t)
		txConfig := clientMocks.NewMockTxConfig(t)

		utilWrapTxBuilder = func(string, string) (client.TxBuilder, client.TxConfig, error) {
			return txBuilder, txConfig, nil
		}

		txBuilder.EXPECT().SetSignatures( mock.Anything).Return(nil)
		txBuilder.EXPECT().GetTx().Return(mock.Anything)
		txConfig.EXPECT().TxJSONEncoder().Return(func(tx sdk.Tx) ([]byte, error) {
			return []byte("{}"), nil
		})

		mockClient.EXPECT().BroadcastTx", mock.Anything).Return("txHash( nil)
		mockDB.EXPECT().UpdateRefund( refund.ID, mock.Anything).Return(nil)

		result := signer.BroadcastRefund(nil, refund, []byte("spender"), sdk.Coin{})

		mockClient.AssertExpectations(t)
		mockDB.AssertExpectations(t)
		txBuilder.AssertExpectations(t)
		txConfig.AssertExpectations(t)
		assert.True(t, result)
	}

	func TestBroadcastRefunds(t *testing.T) {
		mockDB := dbMocks.NewMockDB(t)
		mockClient := clientMocks.NewMockCosmosClient(t)
		logger := log.New().WithField("test", "signer")

		refund := &models.Refund{
			ID:                    &primitive.ObjectID{},
			OriginTransactionHash: "hash1",
			Recipient:             "recipient",
			Amount:                "100",
		}

		signer := &CosmosMessageSignerRunnable{
			db:     mockDB,
			client: mockClient,
			logger: logger,
		}

		mockDB.EXPECT().GetSignedRefunds( mock.Anything).Return([]models.Refund{*refund}, nil)
		mockDB.EXPECT().UpdateRefund( refund.ID, mock.Anything).Return(nil)

		utilValidateTxToCosmosMultisig = func(*sdk.TxResponse, models.CosmosNetworkConfig, map[uint32]bool, uint64) (*util.ValidateTxResult, error) {
			return &util.ValidateTxResult{
				Confirmations: 2,
				TxStatus:      models.TransactionStatusConfirmed,
			}, nil
		}

		result := signer.BroadcastRefunds()

		mockDB.AssertExpectations(t)
		assert.True(t, result)
	}
*/

func TestValidateSignatures(t *testing.T) {
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	txBuilder := clientMocks.NewMockTxBuilder(t)
	txConfig := clientMocks.NewMockTxConfig(t)

	signerKey := secp256k1.GenPrivKey()
	multisigPk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{signerKey.PubKey()})

	signer := &CosmosMessageSignerRunnable{
		client:     mockClient,
		logger:     logger,
		multisigPk: multisigPk,
	}

	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	tx := clientMocks.NewMockTx(t)

	txBuilder.EXPECT().GetTx().Return(tx)
	tx.EXPECT().GetSignaturesV2().Return([]signingtypes.SignatureV2{}, nil)

	txBuilder.EXPECT().SetSignatures(mock.Anything).Return(nil)

	result := signer.ValidateSignatures("hash1", 1, txConfig, txBuilder)

	mockClient.AssertExpectations(t)
	txBuilder.AssertExpectations(t)
	txConfig.AssertExpectations(t)
	assert.True(t, result)
}
func TestFindMaxSequence(t *testing.T) {
	mockDB := dbMocks.NewMockDB(t)
	mockClient := clientMocks.NewMockCosmosClient(t)
	logger := log.New().WithField("test", "signer")

	signer := &CosmosMessageSignerRunnable{
		db:     mockDB,
		client: mockClient,
		logger: logger,
	}

	mockDB.EXPECT().LockReadSequences().Return("lock-id", nil)
	mockDB.EXPECT().Unlock("lock-id").Return(nil)
	mockDB.EXPECT().FindMaxSequence(mock.Anything).Return(nil, nil)
	mockClient.EXPECT().GetAccount(mock.Anything).Return(&authtypes.BaseAccount{AccountNumber: 1, Sequence: 1}, nil)

	sequence, err := signer.FindMaxSequence()

	mockDB.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), sequence)
}
