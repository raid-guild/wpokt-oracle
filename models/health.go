package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CollectionHealthChecks = "healthchecks"
)

type Health struct {
	Id             *primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	CosmosAddress  string              `bson:"cosmos_address" json:"cosmos_address"`
	EthAddress     string              `bson:"eth_address" json:"eth_address"`
	Hostname       string              `bson:"hostname" json:"hostname"`
	CreatedAt      time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time           `bson:"updated_at" json:"updated_at"`
	ServiceHealths []ServiceHealth     `bson:"service_healths" json:"service_healths"`
}

type ChainType string

const (
	ChainTypeEthereum ChainType = "ethereum"
	ChainTypeCosmos   ChainType = "cosmos"
)

type Chain struct {
	ChainId   string    `bson:"chain_id" json:"chain_id"`
	ChainName string    `bson:"chain_name" json:"chain_name"`
	ChainType ChainType `bson:"chain_type" json:"chain_type"`
}

type ServiceHealth struct {
	Chain          Chain        `bson:"chain" json:"chain"`
	MessageMonitor RunnerStatus `bson:"message_monitor" json:"message_monitor"`
	MessageSigner  RunnerStatus `bson:"message_signer" json:"message_signer"`
	MessageRelayer RunnerStatus `bson:"message_relayer" json:"message_relayer"`
}

type RunnerStatus struct {
	Name        string    `bson:"name" json:"name"`
	Enabled     bool      `bson:"enabled" json:"enabled"`
	BlockHeight int64     `bson:"block_height" json:"block_height"`
	LastRunAt   time.Time `bson:"last_run_at" json:"last_run_at"`
	NextRunAt   time.Time `bson:"next_run_at" json:"next_run_at"`
}
