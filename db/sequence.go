package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

type ResultMaxSequence struct {
	MaxSequence uint64 `bson:"max_sequence"`
}

func FindMaxSequenceFromRefunds() (*uint64, error) {
	filter := bson.M{"sequence": bson.M{"$ne": nil}}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "max_sequence", Value: bson.D{{Key: "$max", Value: "$sequence"}}},
		}}},
	}

	var result ResultMaxSequence
	err := MongoDB.AggregateOne(common.CollectionRefunds, pipeline, &result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	maxSequence := uint64(result.MaxSequence)

	return &maxSequence, nil
}

func FindMaxSequenceFromMessages(chain models.Chain) (*uint64, error) {
	filter := bson.M{"content.destination_domain": chain.ChainDomain, "sequence": bson.M{"$ne": nil}}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "max_sequence", Value: bson.D{{Key: "$max", Value: "$sequence"}}},
		}}},
	}

	var result ResultMaxSequence
	err := MongoDB.AggregateOne(common.CollectionMessages, pipeline, &result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	maxSequence := uint64(result.MaxSequence)

	return &maxSequence, nil
}

func FindMaxSequence(chain models.Chain) (*uint64, error) {
	maxSequenceRefunds, err := FindMaxSequenceFromRefunds()
	if err != nil {
		return nil, err
	}

	maxSequenceMessages, err := FindMaxSequenceFromMessages(chain)
	if err != nil {
		return nil, err
	}

	if maxSequenceRefunds == nil && maxSequenceMessages == nil {
		return nil, nil
	}

	if maxSequenceRefunds == nil {
		return maxSequenceMessages, nil
	}

	if maxSequenceMessages == nil {
		return maxSequenceRefunds, nil
	}

	if *maxSequenceRefunds > *maxSequenceMessages {
		return maxSequenceRefunds, nil
	}

	return maxSequenceMessages, nil
}
