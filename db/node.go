package db

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

func FindNode(filter interface{}) (*models.Node, error) {
	var node models.Node
	err := MongoDB.FindOne(common.CollectionNodes, filter, &node)
	return &node, err
}

func UpsertNode(filter interface{}, onUpdate interface{}, onInsert interface{}) error {
	update := bson.M{"$set": onUpdate, "$setOnInsert": onInsert}

	_, err := MongoDB.UpsertOne(common.CollectionNodes, filter, update)
	return err
}
