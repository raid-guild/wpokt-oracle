package db

import (
	"testing"

	mocks "github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LockTestSuite struct {
	suite.Suite
	oldMongoDB Database
	mockDB     *mocks.MockDatabase
}

var oldMongoDB Database

func (suite *LockTestSuite) SetupTest() {
	suite.mockDB = mocks.NewMockDatabase(suite.T())
	suite.oldMongoDB = mongoDB
	mongoDB = suite.mockDB
}

func (suite *LockTestSuite) TearDownTest() {
	mongoDB = suite.oldMongoDB
}

func (suite *LockTestSuite) TestUnlock() {
	lockID := "lock123"
	suite.mockDB.On("Unlock", lockID).Return(nil).Once()

	err := Unlock(lockID)
	assert.NoError(suite.T(), err)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockWriteTransaction() {
	txDoc := &models.Transaction{ID: &primitive.ObjectID{}}
	resourceID := "transactions/" + txDoc.ID.Hex()
	lockID := "lock123"

	suite.mockDB.On("XLock", resourceID).Return(lockID, nil).Once()

	gotLockID, err := LockWriteTransaction(txDoc)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockWriteRefund() {
	refundDoc := &models.Refund{ID: &primitive.ObjectID{}}
	resourceID := "refunds/" + refundDoc.ID.Hex()
	lockID := "lock123"

	suite.mockDB.On("XLock", resourceID).Return(lockID, nil).Once()

	gotLockID, err := LockWriteRefund(refundDoc)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockWriteMessage() {
	messageDoc := &models.Message{ID: &primitive.ObjectID{}}
	resourceID := "messages/" + messageDoc.ID.Hex()
	lockID := "lock123"

	suite.mockDB.On("XLock", resourceID).Return(lockID, nil).Once()

	gotLockID, err := LockWriteMessage(messageDoc)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockReadSequences() {
	lockID := "lock123"
	sequenceResourceID := "comsos_sequence"

	suite.mockDB.On("SLock", sequenceResourceID).Return(lockID, nil).Once()

	gotLockID, err := LockReadSequences()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockWriteSequence() {
	lockID := "lock123"
	sequenceResourceID := "comsos_sequence"

	suite.mockDB.On("SLock", sequenceResourceID).Return(lockID, nil).Once()

	gotLockID, err := LockWriteSequence()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func TestLockTestSuite(t *testing.T) {
	suite.Run(t, new(LockTestSuite))
}
