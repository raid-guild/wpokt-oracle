package db

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

type NodeDB interface {
	FindNode(filter interface{}) (*models.Node, error)
	UpsertNode(filter interface{}, onUpdate interface{}, onInsert interface{}) error
}

func findNode(filter interface{}) (*models.Node, error) {
	var node models.Node
	err := mongoDB.FindOne(common.CollectionNodes, filter, &node)
	return &node, err
}

func upsertNode(filter interface{}, onUpdate interface{}, onInsert interface{}) error {
	update := bson.M{"$set": onUpdate, "$setOnInsert": onInsert}

	_, err := mongoDB.UpsertOne(common.CollectionNodes, filter, update)
	return err
}

type nodeDB struct{}

func (db *nodeDB) FindNode(filter interface{}) (*models.Node, error) {
	return findNode(filter)
}

func (db *nodeDB) UpsertNode(filter interface{}, onUpdate interface{}, onInsert interface{}) error {
	return upsertNode(filter, onUpdate, onInsert)
}
