package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

type ResultMaxSequence struct {
	MaxSequence int `bson:"max_sequence"`
}

func FindMaxSequenceFromRefunds() (uint64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "max_sequence", Value: bson.D{{Key: "$max", Value: "$sequence"}}},
		}}},
	}

	var result ResultMaxSequence
	err := mongoDB.AggregateOne(common.CollectionRefunds, pipeline, &result)
	if err != nil {
		return 0, err
	}

	return uint64(result.MaxSequence), nil
}

func FindMaxSequenceFromMessages(chain models.Chain) (uint64, error) {
	filter := bson.M{"chain": chain}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "max_sequence", Value: bson.D{{Key: "$max", Value: "$sequence"}}},
		}}},
	}

	var result ResultMaxSequence
	err := mongoDB.AggregateOne(common.CollectionMessages, pipeline, &result)
	if err != nil {
		return 0, err
	}

	return uint64(result.MaxSequence), nil
}

func FindMaxSequence(chain models.Chain) (uint64, error) {
	maxSequenceRefunds, err := FindMaxSequenceFromRefunds()
	if err != nil {
		return 0, err
	}

	maxSequenceMessages, err := FindMaxSequenceFromMessages(chain)
	if err != nil {
		return 0, err
	}

	if maxSequenceRefunds > maxSequenceMessages {
		return maxSequenceRefunds, nil
	}

	return maxSequenceMessages, nil
}
