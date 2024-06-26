package db

import (
	"errors"
	"testing"

	"github.com/dan13ram/wpokt-oracle/common"
	mocks "github.com/dan13ram/wpokt-oracle/db/mocks"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NodeTestSuite struct {
	suite.Suite
	mockDB     *mocks.MockDatabase
	oldMongoDB Database
	db         NodeDB
}

func (suite *NodeTestSuite) SetupTest() {
	suite.mockDB = mocks.NewMockDatabase(suite.T())
	suite.oldMongoDB = mongoDB
	mongoDB = suite.mockDB
	suite.db = &nodeDB{}
}

func (suite *NodeTestSuite) TearDownTest() {
	mongoDB = suite.oldMongoDB
}

func (suite *NodeTestSuite) TestFindNode() {
	filter := bson.M{"_id": "some-node-id"}
	expectedNode := models.Node{}

	suite.mockDB.EXPECT().FindOne(common.CollectionNodes, filter, &expectedNode).Return(nil).Once()

	gotNode, err := suite.db.FindNode(filter)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), &expectedNode, gotNode)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *NodeTestSuite) TestFindNode_SomeError() {
	filter := bson.M{"_id": "some-node-id"}
	expectedNode := models.Node{}
	expectedError := errors.New("some error")

	suite.mockDB.EXPECT().FindOne(common.CollectionNodes, filter, &expectedNode).Return(expectedError).Once()

	gotNode, err := suite.db.FindNode(filter)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.Equal(suite.T(), &expectedNode, gotNode)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *NodeTestSuite) TestUpsertNode() {
	filter := bson.M{"_id": "some-node-id"}
	onUpdate := bson.M{"fieldToUpdate": "updatedValue"}
	onInsert := bson.M{"fieldToInsert": "insertedValue"}
	update := bson.M{"$set": onUpdate, "$setOnInsert": onInsert}

	suite.mockDB.EXPECT().UpsertOne(common.CollectionNodes, filter, update).Return(primitive.ObjectID{}, nil).Once()

	err := suite.db.UpsertNode(filter, onUpdate, onInsert)
	assert.NoError(suite.T(), err)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *NodeTestSuite) TestUpsertNode_SomeError() {
	filter := bson.M{"_id": "some-node-id"}
	onUpdate := bson.M{"fieldToUpdate": "updatedValue"}
	onInsert := bson.M{"fieldToInsert": "insertedValue"}
	update := bson.M{"$set": onUpdate, "$setOnInsert": onInsert}
	expectedError := errors.New("some error")

	suite.mockDB.EXPECT().UpsertOne(common.CollectionNodes, filter, update).Return(primitive.ObjectID{}, expectedError).Once()

	err := suite.db.UpsertNode(filter, onUpdate, onInsert)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	suite.mockDB.AssertExpectations(suite.T())
}

func TestNodeTestSuite(t *testing.T) {
	suite.Run(t, new(NodeTestSuite))
}
