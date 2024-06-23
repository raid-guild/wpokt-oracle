package ethereum

import (
	"errors"
	"testing"

	clientMocks "github.com/dan13ram/wpokt-oracle/ethereum/client/mocks"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestValidateTransactionByHash(t *testing.T) {
	mockClient := clientMocks.NewMockEthereumClient(t)
	txHash := "0x123"
	ethTx := &types.Transaction{}
	receipt := &types.Receipt{Status: types.ReceiptStatusSuccessful}

	t.Run("success", func(t *testing.T) {
		mockClient.EXPECT().GetTransactionByHash(txHash).Return(ethTx, false, nil).Once()
		mockClient.EXPECT().GetTransactionReceipt(txHash).Return(receipt, nil).Once()

		result, err := ValidateTransactionByHash(mockClient, txHash)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, ethTx, result.Tx)
		assert.Equal(t, receipt, result.Receipt)
	})

	t.Run("error getting transaction by hash", func(t *testing.T) {
		mockClient.EXPECT().GetTransactionByHash(txHash).Return(nil, false, errors.New("error")).Once()

		result, err := ValidateTransactionByHash(mockClient, txHash)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("transaction not found", func(t *testing.T) {
		mockClient.EXPECT().GetTransactionByHash(txHash).Return(nil, false, nil).Once()

		result, err := ValidateTransactionByHash(mockClient, txHash)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("transaction is pending", func(t *testing.T) {
		mockClient.EXPECT().GetTransactionByHash(txHash).Return(ethTx, true, nil).Once()

		result, err := ValidateTransactionByHash(mockClient, txHash)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error getting transaction receipt", func(t *testing.T) {
		mockClient.EXPECT().GetTransactionByHash(txHash).Return(ethTx, false, nil).Once()
		mockClient.EXPECT().GetTransactionReceipt(txHash).Return(nil, errors.New("error")).Once()

		result, err := ValidateTransactionByHash(mockClient, txHash)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("transaction failed", func(t *testing.T) {
		failedReceipt := &types.Receipt{Status: types.ReceiptStatusFailed}
		mockClient.EXPECT().GetTransactionByHash(txHash).Return(ethTx, false, nil).Once()
		mockClient.EXPECT().GetTransactionReceipt(txHash).Return(failedReceipt, nil).Once()

		result, err := ValidateTransactionByHash(mockClient, txHash)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
