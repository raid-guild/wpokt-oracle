package db

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dan13ram/wpokt-oracle/common"
	mocks "github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/models"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionTestSuite struct {
	suite.Suite
	mockDB     *mocks.MockDatabase
	oldMongoDB Database
}

func (suite *TransactionTestSuite) SetupTest() {
	suite.mockDB = mocks.NewMockDatabase(suite.T())
	suite.oldMongoDB = mongoDB
	mongoDB = suite.mockDB
}

func (suite *TransactionTestSuite) TearDownTest() {
	mongoDB = suite.oldMongoDB
}

func (suite *TransactionTestSuite) TestNewEthereumTransaction() {
	testAddr := ethcommon.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	tx, err := types.NewTx(&types.AccessListTx{
		ChainID:  big.NewInt(1),
		Nonce:    3,
		To:       &testAddr,
		Value:    big.NewInt(10),
		Gas:      25000,
		GasPrice: big.NewInt(1),
		Data:     ethcommon.FromHex("5544"),
	}).WithSignature(
		types.NewEIP2930Signer(big.NewInt(1)),
		ethcommon.Hex2Bytes("c9519f4f2b30335884581971573fadf60c6204f59a911df35ee8a540456b266032f1e8e2c5dd761f9e4f88f41c8310aeaba26a8bfcdacfedfa12ec3862d3752101"),
	)
	if err != nil {
		panic(err)
	}
	toAddress := ethcommon.HexToAddress("0x")
	receipt := &types.Receipt{
		TxHash:      ethcommon.HexToHash("0x01"),
		BlockNumber: big.NewInt(1),
	}
	chain := models.Chain{ChainID: "eth"}
	txStatus := models.TransactionStatusPending

	ethTx, err := NewEthereumTransaction(tx, toAddress.Bytes(), receipt, chain, txStatus)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.Ensure0xPrefix(receipt.TxHash.String()), ethTx.Hash)
	assert.Equal(suite.T(), toAddress.Hex(), ethTx.ToAddress)
	assert.Equal(suite.T(), chain, ethTx.Chain)
	assert.Equal(suite.T(), txStatus, ethTx.Status)
}

func (suite *TransactionTestSuite) TestNewEthereumTransaction_InvalidToAddress() {
	testAddr := ethcommon.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	tx, err := types.NewTx(&types.AccessListTx{
		ChainID:  big.NewInt(1),
		Nonce:    3,
		To:       &testAddr,
		Value:    big.NewInt(10),
		Gas:      25000,
		GasPrice: big.NewInt(1),
		Data:     ethcommon.FromHex("5544"),
	}).WithSignature(
		types.NewEIP2930Signer(big.NewInt(1)),
		ethcommon.Hex2Bytes("c9519f4f2b30335884581971573fadf60c6204f59a911df35ee8a540456b266032f1e8e2c5dd761f9e4f88f41c8310aeaba26a8bfcdacfedfa12ec3862d3752101"),
	)
	if err != nil {
		panic(err)
	}
	toAddress := []byte{0x01}
	receipt := &types.Receipt{
		TxHash:      ethcommon.HexToHash("0x01"),
		BlockNumber: big.NewInt(1),
	}
	chain := models.Chain{ChainID: "eth"}
	txStatus := models.TransactionStatusPending
	expectedError := fmt.Errorf("invalid to address: %w", common.ErrInvalidAddressLength)

	_, err = NewEthereumTransaction(tx, toAddress, receipt, chain, txStatus)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

func (suite *TransactionTestSuite) TestNewEthereumTransaction_InvalidFromAddress() {
	testAddr := ethcommon.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	tx := types.NewTx(&types.AccessListTx{
		ChainID:  big.NewInt(1),
		Nonce:    3,
		To:       &testAddr,
		Value:    big.NewInt(10),
		Gas:      25000,
		GasPrice: big.NewInt(1),
		Data:     ethcommon.FromHex("5544"),
	})
	toAddress := ethcommon.HexToAddress("0x")
	receipt := &types.Receipt{
		TxHash:      ethcommon.HexToHash("0x01"),
		BlockNumber: big.NewInt(1),
	}
	chain := models.Chain{ChainID: "eth"}
	txStatus := models.TransactionStatusPending

	_, err := NewEthereumTransaction(tx, toAddress.Bytes(), receipt, chain, txStatus)
	assert.Error(suite.T(), err)
}

func (suite *TransactionTestSuite) TestNewCosmosTransaction() {
	txHash := ethcommon.HexToHash("0x01020304")
	txRes := &sdk.TxResponse{TxHash: txHash.Hex(), Height: 1}
	fromAddress := ethcommon.HexToAddress("0x010203")
	toAddress := ethcommon.HexToAddress("0x040506")
	chain := models.Chain{ChainID: "cosmos"}
	txStatus := models.TransactionStatusPending

	cosmosTx, err := NewCosmosTransaction(txRes, chain, fromAddress[:], toAddress[:], txStatus)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.Ensure0xPrefix(txRes.TxHash), cosmosTx.Hash)
	assert.Equal(suite.T(), strings.ToLower(fromAddress.Hex()), cosmosTx.FromAddress)
	assert.Equal(suite.T(), strings.ToLower(toAddress.Hex()), cosmosTx.ToAddress)
	assert.Equal(suite.T(), chain, cosmosTx.Chain)
	assert.Equal(suite.T(), txStatus, cosmosTx.Status)
}

func (suite *TransactionTestSuite) TestNewCosmosTransaction_InvalidFromAddress() {
	txHash := ethcommon.HexToHash("0x01020304")
	txRes := &sdk.TxResponse{TxHash: txHash.Hex(), Height: 1}
	fromAddress := []byte{0x01}
	toAddress := ethcommon.HexToAddress("0x040506")
	chain := models.Chain{ChainID: "cosmos"}
	txStatus := models.TransactionStatusPending
	expectedError := fmt.Errorf("invalid from address: %w", common.ErrInvalidAddressLength)

	_, err := NewCosmosTransaction(txRes, chain, fromAddress[:], toAddress[:], txStatus)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

func (suite *TransactionTestSuite) TestNewCosmosTransaction_InvalidToAddress() {
	txHash := ethcommon.HexToHash("0x01020304")
	txRes := &sdk.TxResponse{TxHash: txHash.Hex(), Height: 1}
	toAddress := []byte{0x01}
	fromAddress := ethcommon.HexToAddress("0x040506")
	chain := models.Chain{ChainID: "cosmos"}
	txStatus := models.TransactionStatusPending
	expectedError := fmt.Errorf("invalid to address: %w", common.ErrInvalidAddressLength)

	_, err := NewCosmosTransaction(txRes, chain, fromAddress[:], toAddress[:], txStatus)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

func (suite *TransactionTestSuite) TestNewCosmosTransaction_InvalidTxHash() {
	txRes := &sdk.TxResponse{TxHash: "0x01020304", Height: 1}
	toAddress := ethcommon.HexToAddress("0x010203")
	fromAddress := ethcommon.HexToAddress("0x040506")
	chain := models.Chain{ChainID: "cosmos"}
	txStatus := models.TransactionStatusPending
	expectedError := fmt.Errorf("invalid tx hash")

	_, err := NewCosmosTransaction(txRes, chain, fromAddress[:], toAddress[:], txStatus)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

func (suite *TransactionTestSuite) TestInsertTransaction() {
	tx := models.Transaction{Hash: "01020304"}
	insertedID := primitive.NewObjectID()

	suite.mockDB.On("InsertOne", common.CollectionTransactions, tx).Return(insertedID, nil).Once()

	gotID, err := InsertTransaction(tx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), insertedID, gotID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *TransactionTestSuite) TestInsertTransaction_DuplicateKeyError() {
	tx := models.Transaction{Hash: "01020304"}
	duplicateError := mongo.WriteError{Code: 11000}
	insertedID := primitive.NewObjectID()
	existingTx := models.Transaction{ID: &insertedID}

	suite.mockDB.On("InsertOne", common.CollectionTransactions, tx).Return(primitive.ObjectID{}, duplicateError).Once()
	suite.mockDB.On("FindOne", common.CollectionTransactions, bson.M{"hash": tx.Hash}, &models.Transaction{}).Return(nil).Once().Run(func(args mock.Arguments) {
		tx := args.Get(2).(*models.Transaction)
		*tx = existingTx
	})

	gotID, err := InsertTransaction(tx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), insertedID, gotID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *TransactionTestSuite) TestInsertTransaction_DuplicateKeyError_FindError() {
	tx := models.Transaction{
		Hash: "0x123",
	}
	duplicateError := mongo.WriteError{Code: 11000}
	insertedID := primitive.NewObjectID()
	expectedError := fmt.Errorf("find error")

	suite.mockDB.On("InsertOne", common.CollectionTransactions, tx).Return(insertedID, duplicateError).Once()
	suite.mockDB.On("FindOne", common.CollectionTransactions, bson.M{"hash": tx.Hash}, &models.Transaction{}).Return(expectedError).Once()

	gotID, err := InsertTransaction(tx)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.Equal(suite.T(), insertedID, gotID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *TransactionTestSuite) TestInsertTransaction_InsertError() {
	tx := models.Transaction{
		Hash: "0x123",
	}
	insertedID := primitive.NewObjectID()
	expectedError := fmt.Errorf("insert error")

	suite.mockDB.On("InsertOne", common.CollectionTransactions, tx).Return(insertedID, expectedError).Once()

	gotID, err := InsertTransaction(tx)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.Equal(suite.T(), insertedID, gotID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *TransactionTestSuite) TestUpdateTransaction() {
	txID := primitive.NewObjectID()
	update := bson.M{"status": models.TransactionStatusConfirmed}

	suite.mockDB.On("UpdateOne", common.CollectionTransactions, bson.M{"_id": txID}, bson.M{"$set": update}).Return(primitive.ObjectID{}, nil).Once()

	err := UpdateTransaction(&txID, update)
	assert.NoError(suite.T(), err)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *TransactionTestSuite) TestUpdateTransaction_NilTxDoc() {
	update := bson.M{"status": models.TransactionStatusConfirmed}

	err := UpdateTransaction(nil, update)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "txID is nil", err.Error())
}

func (suite *TransactionTestSuite) TestGetPendingTransactionsTo() {
	chain := models.Chain{ChainID: "eth"}
	toAddress := ethcommon.HexToAddress("0x010203")
	expectedTxs := []models.Transaction{
		{ID: &primitive.ObjectID{}, Status: models.TransactionStatusPending},
	}

	filter := bson.M{
		"status":     models.TransactionStatusPending,
		"chain":      chain,
		"to_address": strings.ToLower(toAddress.Hex()),
	}

	suite.mockDB.On("FindMany", common.CollectionTransactions, filter, &[]models.Transaction{}).Return(nil).Once().Run(func(args mock.Arguments) {
		txs := args.Get(2).(*[]models.Transaction)
		*txs = expectedTxs
	})

	gotTxs, err := GetPendingTransactionsTo(chain, toAddress[:])
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTxs, gotTxs)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *TransactionTestSuite) TestGetPendingTransactionsTo_InvalidToAddress() {
	chain := models.Chain{ChainID: "eth"}
	toAddress := []byte{0x01}

	expectedError := fmt.Errorf("invalid to address: %w", common.ErrInvalidAddressLength)
	_, err := GetPendingTransactionsTo(chain, toAddress[:])
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

func (suite *TransactionTestSuite) TestGetConfirmedTransactionsTo() {
	chain := models.Chain{ChainID: "eth"}
	toAddress := ethcommon.HexToAddress("0x010203")
	expectedTxs := []models.Transaction{
		{ID: &primitive.ObjectID{}, Status: models.TransactionStatusConfirmed},
	}

	filter := bson.M{
		"$and": []bson.M{
			{
				"status":     models.TransactionStatusConfirmed,
				"chain":      chain,
				"to_address": strings.ToLower(toAddress.Hex()),
			},
			{"$or": []bson.M{
				{"refund": bson.M{"$exists": false}},
				{"refund": bson.M{"$eq": nil}},
			}},
			{"$or": []bson.M{
				{"messages": bson.M{"$exists": false}},
				{"messages": bson.M{"$eq": nil}},
				{"messages": bson.M{"$size": 0}},
			}},
		},
	}

	suite.mockDB.On("FindMany", common.CollectionTransactions, filter, &[]models.Transaction{}).Return(nil).Once().Run(func(args mock.Arguments) {
		txs := args.Get(2).(*[]models.Transaction)
		*txs = expectedTxs
	})

	gotTxs, err := GetConfirmedTransactionsTo(chain, toAddress[:])
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTxs, gotTxs)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *TransactionTestSuite) TestGetConfirmedTransactionsTo_InvalidToAddress() {
	chain := models.Chain{ChainID: "eth"}
	toAddress := []byte{0x01}

	expectedError := fmt.Errorf("invalid to address: %w", common.ErrInvalidAddressLength)
	_, err := GetConfirmedTransactionsTo(chain, toAddress[:])
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

func (suite *TransactionTestSuite) TestGetPendingTransactionsFrom() {
	chain := models.Chain{ChainID: "eth"}
	fromAddress := ethcommon.HexToAddress("0x010203")
	expectedTxs := []models.Transaction{
		{ID: &primitive.ObjectID{}, Status: models.TransactionStatusPending},
	}

	filter := bson.M{
		"status":       models.TransactionStatusPending,
		"chain":        chain,
		"from_address": strings.ToLower(fromAddress.Hex()),
	}

	suite.mockDB.On("FindMany", common.CollectionTransactions, filter, &[]models.Transaction{}).Return(nil).Once().Run(func(args mock.Arguments) {
		txs := args.Get(2).(*[]models.Transaction)
		*txs = expectedTxs
	})

	gotTxs, err := GetPendingTransactionsFrom(chain, fromAddress[:])
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTxs, gotTxs)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *TransactionTestSuite) TestGetPendingTransactionsFrom_InvalidFromAddress() {
	chain := models.Chain{ChainID: "eth"}
	fromAddress := []byte{0x01}

	expectedError := fmt.Errorf("invalid from address: %w", common.ErrInvalidAddressLength)
	_, err := GetPendingTransactionsFrom(chain, fromAddress[:])
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

func TestTransactionTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionTestSuite))
}
