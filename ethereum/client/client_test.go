package client

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
)

func TestEthereumClient_GetBlockHeight(t *testing.T) {
	mockClient := NewMockEthHTTPClient(t)
	ethClient := &ethereumClient{
		client:  mockClient,
		timeout: 5 * time.Second,
	}

	mockClient.On("BlockNumber", mock.Anything).Return(uint64(12345), nil)

	blockHeight, err := ethClient.GetBlockHeight()
	assert.NoError(t, err)
	assert.Equal(t, uint64(12345), blockHeight)

	mockClient.AssertExpectations(t)
}

func TestEthereumClient_Chain(t *testing.T) {
	ethClient := &ethereumClient{}
	chain := ethClient.Chain()
	assert.Equal(t, models.Chain{}, chain)
}

func TestEthereumClient_GetClient(t *testing.T) {
	mockClient := NewMockEthHTTPClient(t)
	ethClient := &ethereumClient{
		client:  mockClient,
		timeout: 5 * time.Second,
	}

	client := ethClient.GetClient()
	assert.Equal(t, mockClient, client)
}

func TestEthereumClient_GetBlockHeight_Error(t *testing.T) {
	mockClient := NewMockEthHTTPClient(t)
	ethClient := &ethereumClient{
		client:  mockClient,
		timeout: 5 * time.Second,
	}

	expectedError := fmt.Errorf("failed to get block number")
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(12345), expectedError)

	blockHeight, err := ethClient.GetBlockHeight()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError.Error())
	assert.Equal(t, uint64(0), blockHeight)

	mockClient.AssertExpectations(t)
}

func TestEthereumClient_GetChainID(t *testing.T) {
	mockClient := NewMockEthHTTPClient(t)
	ethClient := &ethereumClient{
		client:  mockClient,
		timeout: 5 * time.Second,
	}

	expectedChainID := big.NewInt(1)
	mockClient.On("ChainID", mock.Anything).Return(expectedChainID, nil)

	chainID, err := ethClient.GetChainID()
	assert.NoError(t, err)
	assert.Equal(t, expectedChainID, chainID)

	mockClient.AssertExpectations(t)
}

func TestEthereumClient_ValidateNetwork(t *testing.T) {
	mockClient := NewMockEthHTTPClient(t)
	ethClient := &ethereumClient{
		client:  mockClient,
		timeout: 5 * time.Second,
		chainID: 1,
		logger:  log.NewEntry(log.New()),
	}

	expectedChainID := big.NewInt(1)
	mockClient.On("ChainID", mock.Anything).Return(expectedChainID, nil)

	err := ethClient.ValidateNetwork()
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestEthereumClient_ValidateNetwork_ErrorChainID(t *testing.T) {
	mockClient := NewMockEthHTTPClient(t)
	ethClient := &ethereumClient{
		client:  mockClient,
		timeout: 5 * time.Second,
		chainID: 1,
		logger:  log.NewEntry(log.New()),
	}

	expectedError := fmt.Errorf("failed to get chain id")
	mockClient.On("ChainID", mock.Anything).Return(nil, expectedError)

	err := ethClient.ValidateNetwork()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError.Error())

	mockClient.AssertExpectations(t)
}

func TestEthereumClient_ValidateNetwork_InvalidChainID(t *testing.T) {
	mockClient := NewMockEthHTTPClient(t)
	ethClient := &ethereumClient{
		client:  mockClient,
		timeout: 5 * time.Second,
		chainID: 1,
		logger:  log.NewEntry(log.New()),
	}

	wrongChainID := big.NewInt(2)
	mockClient.On("ChainID", mock.Anything).Return(wrongChainID, nil)

	err := ethClient.ValidateNetwork()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected chain id 1, got 2")

	mockClient.AssertExpectations(t)
}

func TestEthereumClient_GetTransactionByHash(t *testing.T) {
	mockClient := NewMockEthHTTPClient(t)
	ethClient := &ethereumClient{
		client:  mockClient,
		timeout: 5 * time.Second,
	}

	txHash := common.HexToHash("0x123")
	expectedTx := &types.Transaction{}
	mockClient.On("TransactionByHash", mock.Anything, txHash).Return(expectedTx, false, nil)

	tx, isPending, err := ethClient.GetTransactionByHash("0x123")
	assert.NoError(t, err)
	assert.Equal(t, expectedTx, tx)
	assert.False(t, isPending)

	mockClient.AssertExpectations(t)
}

func TestEthereumClient_GetTransactionReceipt(t *testing.T) {
	mockClient := NewMockEthHTTPClient(t)
	ethClient := &ethereumClient{
		client:  mockClient,
		timeout: 5 * time.Second,
	}

	txHash := common.HexToHash("0x123")
	expectedReceipt := &types.Receipt{}
	mockClient.On("TransactionReceipt", mock.Anything, txHash).Return(expectedReceipt, nil)

	receipt, err := ethClient.GetTransactionReceipt("0x123")
	assert.NoError(t, err)
	assert.Equal(t, expectedReceipt, receipt)

	mockClient.AssertExpectations(t)
}

func TestNewClient(t *testing.T) {
	config := models.EthereumNetworkConfig{
		RPCURL:        "http://localhost:8545",
		TimeoutMS:     5000,
		ChainID:       1,
		ChainName:     "Ethereum",
		Confirmations: 12,
	}

	mockClient := NewMockEthHTTPClient(t)

	mockClient.On("ChainID", mock.Anything).Return(big.NewInt(int64(config.ChainID)), nil)

	oldEthclientDial := ethclientDial
	ethclientDial = func(url string) (EthHTTPClient, error) {
		return mockClient, nil
	}
	defer func() { ethclientDial = oldEthclientDial }()

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, uint64(12), client.Confirmations())
}

func TestNewClient_InvalidRPC(t *testing.T) {
	config := models.EthereumNetworkConfig{
		RPCURL:        "invalid-url",
		TimeoutMS:     5000,
		ChainID:       1,
		ChainName:     "Ethereum",
		Confirmations: 12,
	}

	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestNewClient_ConnectionFailed(t *testing.T) {
	config := models.EthereumNetworkConfig{
		RPCURL:        "valid-url",
		TimeoutMS:     5000,
		ChainID:       1,
		ChainName:     "Ethereum",
		Confirmations: 12,
	}

	oldEthclientDial := ethclientDial
	ethclientDial = func(url string) (EthHTTPClient, error) {
		return nil, fmt.Errorf("failed to connect to rpc")
	}
	defer func() { ethclientDial = oldEthclientDial }()

	client, err := NewClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestNewClient_ValidationFailed(t *testing.T) {
	mockClient := NewMockEthHTTPClient(t)

	wrongChainID := big.NewInt(2)
	mockClient.On("ChainID", mock.Anything).Return(wrongChainID, nil)

	config := models.EthereumNetworkConfig{
		RPCURL:        "valid-url",
		TimeoutMS:     5000,
		ChainID:       1,
		ChainName:     "Ethereum",
		Confirmations: 12,
	}

	oldEthclientDial := ethclientDial
	ethclientDial = func(url string) (EthHTTPClient, error) {
		return mockClient, nil
	}
	defer func() { ethclientDial = oldEthclientDial }()

	client, err := NewClient(config)
	assert.Contains(t, err.Error(), "failed to validate network")
	assert.Error(t, err)
	assert.Nil(t, client)
	mockClient.AssertExpectations(t)
}
