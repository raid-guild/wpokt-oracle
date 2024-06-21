package db

import (
	"errors"
	"testing"

	"github.com/dan13ram/wpokt-oracle/common"
	mocks "github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SequenceTestSuite struct {
	suite.Suite
	mockDB     *mocks.MockDatabase
	oldMongoDB Database
	db         SequenceDB
}

func (suite *SequenceTestSuite) SetupTest() {
	suite.mockDB = mocks.NewMockDatabase(suite.T())
	suite.oldMongoDB = mongoDB
	mongoDB = suite.mockDB
	suite.db = &sequenceDB{}
}

func (suite *SequenceTestSuite) TearDownTest() {
	mongoDB = suite.oldMongoDB
}

func (suite *SequenceTestSuite) TestFindMaxSequenceFromRefunds() {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"sequence": bson.M{"$ne": nil}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "max_sequence", Value: bson.D{{Key: "$max", Value: "$sequence"}}},
		}}},
	}

	result := resultMaxSequence{}
	suite.mockDB.EXPECT().AggregateOne(common.CollectionRefunds, pipeline, &result).Return(nil).Once()

	maxSequence, err := findMaxSequenceFromRefunds()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uint64(0), *maxSequence)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *SequenceTestSuite) TestFindMaxSequenceFromRefunds_NoDocuments() {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"sequence": bson.M{"$ne": nil}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "max_sequence", Value: bson.D{{Key: "$max", Value: "$sequence"}}},
		}}},
	}

	var result resultMaxSequence
	suite.mockDB.EXPECT().AggregateOne(common.CollectionRefunds, pipeline, &result).Return(mongo.ErrNoDocuments).Once()

	maxSequence, err := findMaxSequenceFromRefunds()
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), maxSequence)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *SequenceTestSuite) TestFindMaxSequenceFromRefunds_SomeError() {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"sequence": bson.M{"$ne": nil}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "max_sequence", Value: bson.D{{Key: "$max", Value: "$sequence"}}},
		}}},
	}
	expectedError := errors.New("some error")

	suite.mockDB.EXPECT().AggregateOne(common.CollectionRefunds, pipeline, mock.Anything).Return(expectedError).Once()

	maxSequence, err := findMaxSequenceFromRefunds()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.Nil(suite.T(), maxSequence)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *SequenceTestSuite) TestFindMaxSequenceFromMessages() {
	chain := models.Chain{ChainDomain: 1}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"content.destination_domain": chain.ChainDomain, "sequence": bson.M{"$ne": nil}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "max_sequence", Value: bson.D{{Key: "$max", Value: "$sequence"}}},
		}}},
	}

	result := resultMaxSequence{}
	suite.mockDB.EXPECT().AggregateOne(common.CollectionMessages, pipeline, &result).Return(nil).Once()

	maxSequence, err := findMaxSequenceFromMessages(chain)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uint64(0), *maxSequence)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *SequenceTestSuite) TestFindMaxSequenceFromMessages_NoDocuments() {
	chain := models.Chain{ChainDomain: 1}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"content.destination_domain": chain.ChainDomain, "sequence": bson.M{"$ne": nil}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "max_sequence", Value: bson.D{{Key: "$max", Value: "$sequence"}}},
		}}},
	}

	var result resultMaxSequence
	suite.mockDB.EXPECT().AggregateOne(common.CollectionMessages, pipeline, &result).Return(mongo.ErrNoDocuments).Once()

	maxSequence, err := findMaxSequenceFromMessages(chain)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), maxSequence)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *SequenceTestSuite) TestFindMaxSequenceFromMessages_SomeError() {
	chain := models.Chain{ChainDomain: 1}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"content.destination_domain": chain.ChainDomain, "sequence": bson.M{"$ne": nil}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "max_sequence", Value: bson.D{{Key: "$max", Value: "$sequence"}}},
		}}},
	}
	expectedError := errors.New("some error")

	suite.mockDB.EXPECT().AggregateOne(common.CollectionMessages, pipeline, mock.Anything).Return(expectedError).Once()

	maxSequence, err := findMaxSequenceFromMessages(chain)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.Nil(suite.T(), maxSequence)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *SequenceTestSuite) TestFindMaxSequence() {
	chain := models.Chain{ChainDomain: 1}
	maxSequenceRefunds := uint64(123)
	maxSequenceMessages := uint64(456)

	suite.mockDB.EXPECT().AggregateOne(common.CollectionRefunds, mock.Anything, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
		result := args.Get(2).(*resultMaxSequence)
		result.MaxSequence = maxSequenceRefunds
	}).Once()

	suite.mockDB.EXPECT().AggregateOne(common.CollectionMessages, mock.Anything, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
		result := args.Get(2).(*resultMaxSequence)
		result.MaxSequence = maxSequenceMessages
	}).Once()

	maxSequence, err := suite.db.FindMaxSequence(chain)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), maxSequenceMessages, *maxSequence)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *SequenceTestSuite) TestFindMaxSequence_RefundsGreater() {
	chain := models.Chain{ChainDomain: 1}
	maxSequenceRefunds := uint64(456)
	maxSequenceMessages := uint64(123)

	suite.mockDB.EXPECT().AggregateOne(common.CollectionRefunds, mock.Anything, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
		result := args.Get(2).(*resultMaxSequence)
		result.MaxSequence = maxSequenceRefunds
	}).Once()

	suite.mockDB.EXPECT().AggregateOne(common.CollectionMessages, mock.Anything, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
		result := args.Get(2).(*resultMaxSequence)
		result.MaxSequence = maxSequenceMessages
	}).Once()

	maxSequence, err := suite.db.FindMaxSequence(chain)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), maxSequenceRefunds, *maxSequence)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *SequenceTestSuite) TestFindMaxSequence_ErrorMessages() {
	chain := models.Chain{ChainDomain: 1}
	maxSequenceRefunds := uint64(123)
	maxSequenceMessages := uint64(456)
	expectedError := errors.New("some error")

	suite.mockDB.EXPECT().AggregateOne(common.CollectionRefunds, mock.Anything, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
		result := args.Get(2).(*resultMaxSequence)
		result.MaxSequence = maxSequenceRefunds
	}).Once()

	suite.mockDB.EXPECT().AggregateOne(common.CollectionMessages, mock.Anything, mock.Anything).Return(expectedError).Once().Run(func(args mock.Arguments) {
		result := args.Get(2).(*resultMaxSequence)
		result.MaxSequence = maxSequenceMessages
	}).Once()

	_, err := suite.db.FindMaxSequence(chain)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *SequenceTestSuite) TestFindMaxSequence_ErrorRefunds() {
	chain := models.Chain{ChainDomain: 1}
	maxSequenceRefunds := uint64(123)
	expectedError := errors.New("some error")

	suite.mockDB.EXPECT().AggregateOne(common.CollectionRefunds, mock.Anything, mock.Anything).Return(expectedError).Once().Run(func(args mock.Arguments) {
		result := args.Get(2).(*resultMaxSequence)
		result.MaxSequence = maxSequenceRefunds
	}).Once()

	_, err := suite.db.FindMaxSequence(chain)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *SequenceTestSuite) TestFindMaxSequence_OnlyRefunds() {
	chain := models.Chain{ChainDomain: 1}
	maxSequenceRefunds := uint64(123)

	suite.mockDB.EXPECT().AggregateOne(common.CollectionRefunds, mock.Anything, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
		result := args.Get(2).(*resultMaxSequence)
		result.MaxSequence = maxSequenceRefunds
	}).Once()

	suite.mockDB.EXPECT().AggregateOne(common.CollectionMessages, mock.Anything, mock.Anything).Return(mongo.ErrNoDocuments).Once()

	maxSequence, err := suite.db.FindMaxSequence(chain)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), maxSequenceRefunds, *maxSequence)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *SequenceTestSuite) TestFindMaxSequence_OnlyMessages() {
	chain := models.Chain{ChainDomain: 1}
	maxSequenceMessages := uint64(456)

	suite.mockDB.EXPECT().AggregateOne(common.CollectionRefunds, mock.Anything, mock.Anything).Return(mongo.ErrNoDocuments).Once()

	suite.mockDB.EXPECT().AggregateOne(common.CollectionMessages, mock.Anything, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
		result := args.Get(2).(*resultMaxSequence)
		result.MaxSequence = maxSequenceMessages
	}).Once()

	maxSequence, err := suite.db.FindMaxSequence(chain)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), maxSequenceMessages, *maxSequence)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *SequenceTestSuite) TestFindMaxSequence_NoSequences() {
	chain := models.Chain{ChainDomain: 1}

	suite.mockDB.EXPECT().AggregateOne(common.CollectionRefunds, mock.Anything, mock.Anything).Return(mongo.ErrNoDocuments).Once()
	suite.mockDB.EXPECT().AggregateOne(common.CollectionMessages, mock.Anything, mock.Anything).Return(mongo.ErrNoDocuments).Once()

	maxSequence, err := suite.db.FindMaxSequence(chain)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), maxSequence)
	suite.mockDB.AssertExpectations(suite.T())
}

func TestSequenceTestSuite(t *testing.T) {
	suite.Run(t, new(SequenceTestSuite))
}
