package db

import (
	"errors"
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
	db         LockDB
}

func (suite *LockTestSuite) SetupTest() {
	suite.mockDB = mocks.NewMockDatabase(suite.T())
	suite.oldMongoDB = mongoDB
	mongoDB = suite.mockDB
	suite.db = &lockDB{}
}

func (suite *LockTestSuite) TearDownTest() {
	mongoDB = suite.oldMongoDB
}

func (suite *LockTestSuite) TestUnlock() {
	lockID := "lock123"
	suite.mockDB.EXPECT().Unlock(lockID).Return(nil).Once()

	err := suite.db.Unlock(lockID)
	assert.NoError(suite.T(), err)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockWriteTransaction() {
	txDoc := &models.Transaction{ID: &primitive.ObjectID{}}
	resourceID := "transactions/" + txDoc.ID.Hex()
	lockID := "lock123"

	suite.mockDB.EXPECT().XLock(resourceID).Return(lockID, nil).Once()

	gotLockID, err := suite.db.LockWriteTransaction(txDoc)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockWriteTransaction_SomeError() {
	txDoc := &models.Transaction{ID: &primitive.ObjectID{}}
	resourceID := "transactions/" + txDoc.ID.Hex()
	lockID := "lock123"
	expectedErr := errors.New("some error")

	suite.mockDB.EXPECT().XLock(resourceID).Return(lockID, expectedErr).Once()

	gotLockID, err := suite.db.LockWriteTransaction(txDoc)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedErr, err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockWriteRefund() {
	refundDoc := &models.Refund{ID: &primitive.ObjectID{}}
	resourceID := "refunds/" + refundDoc.ID.Hex()
	lockID := "lock123"

	suite.mockDB.EXPECT().XLock(resourceID).Return(lockID, nil).Once()

	gotLockID, err := suite.db.LockWriteRefund(refundDoc)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockWriteRefund_SomeError() {
	refundDoc := &models.Refund{ID: &primitive.ObjectID{}}
	resourceID := "refunds/" + refundDoc.ID.Hex()
	lockID := "lock123"
	expectedErr := errors.New("some error")

	suite.mockDB.EXPECT().XLock(resourceID).Return(lockID, expectedErr).Once()

	gotLockID, err := suite.db.LockWriteRefund(refundDoc)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedErr, err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockWriteMessage() {
	messageDoc := &models.Message{ID: &primitive.ObjectID{}}
	resourceID := "messages/" + messageDoc.ID.Hex()
	lockID := "lock123"

	suite.mockDB.EXPECT().XLock(resourceID).Return(lockID, nil).Once()

	gotLockID, err := suite.db.LockWriteMessage(messageDoc)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockWriteMessage_SomeError() {
	messageDoc := &models.Message{ID: &primitive.ObjectID{}}
	resourceID := "messages/" + messageDoc.ID.Hex()
	lockID := "lock123"
	expectedErr := errors.New("some error")

	suite.mockDB.EXPECT().XLock(resourceID).Return(lockID, expectedErr).Once()

	gotLockID, err := suite.db.LockWriteMessage(messageDoc)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedErr, err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockReadSequences() {
	lockID := "lock123"
	sequenceResourceID := "comsos_sequence"

	suite.mockDB.EXPECT().SLock(sequenceResourceID).Return(lockID, nil).Once()

	gotLockID, err := suite.db.LockReadSequences()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockReadSequences_SomeError() {
	lockID := "lock123"
	sequenceResourceID := "comsos_sequence"
	expectedErr := errors.New("some error")

	suite.mockDB.EXPECT().SLock(sequenceResourceID).Return(lockID, expectedErr).Once()

	gotLockID, err := suite.db.LockReadSequences()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedErr, err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockWriteSequence() {
	lockID := "lock123"
	sequenceResourceID := "comsos_sequence"

	suite.mockDB.EXPECT().SLock(sequenceResourceID).Return(lockID, nil).Once()

	gotLockID, err := suite.db.LockWriteSequence()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *LockTestSuite) TestLockWriteSequence_SomeError() {
	lockID := "lock123"
	sequenceResourceID := "comsos_sequence"
	expectedErr := errors.New("some error")

	suite.mockDB.EXPECT().SLock(sequenceResourceID).Return(lockID, expectedErr).Once()

	gotLockID, err := suite.db.LockWriteSequence()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedErr, err)
	assert.Equal(suite.T(), lockID, gotLockID)
	suite.mockDB.AssertExpectations(suite.T())
}

func TestLockTestSuite(t *testing.T) {
	suite.Run(t, new(LockTestSuite))
}
