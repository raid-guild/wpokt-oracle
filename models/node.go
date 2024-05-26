package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Node struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"_id"`
	CosmosAddress   []byte               `bson:"cosmos_address" json:"cosmos_address"`
	EthAddress      []byte               `bson:"eth_address" json:"eth_address"`
	Hostname        string               `bson:"hostname" json:"hostname"`
	OracleID        string               `bson:"oracle_id" json:"oracle_id"`
	SupportedChains []Chain              `bson:"supported_chains" json:"supported_chains"`
	Health          []ChainServiceHealth `bson:"service_healths" json:"service_healths"`
	CreatedAt       time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time            `bson:"updated_at" json:"updated_at"`
}

type ChainType string

const (
	ChainTypeEthereum ChainType = "ethereum"
	ChainTypeCosmos   ChainType = "cosmos"
)

type Chain struct {
	ChainID     string    `bson:"chain_id" json:"chain_id"`
	ChainName   string    `bson:"chain_name" json:"chain_name"`
	ChainDomain uint32    `bson:"chain_domain" json:"chain_domain"`
	ChainType   ChainType `bson:"chain_type" json:"chain_type"`
}

type ChainServiceHealth struct {
	Chain          Chain                `bson:"chain" json:"chain"`
	MessageMonitor *RunnerServiceStatus `bson:"message_monitor" json:"message_monitor"`
	MessageSigner  *RunnerServiceStatus `bson:"message_signer" json:"message_signer"`
	MessageRelayer *RunnerServiceStatus `bson:"message_relayer" json:"message_relayer"`
}

type RunnerServiceStatus struct {
	Name        string    `bson:"name" json:"name"`
	Enabled     bool      `bson:"enabled" json:"enabled"`
	BlockHeight uint64    `bson:"block_height" json:"block_height"`
	LastRunAt   time.Time `bson:"last_run_at" json:"last_run_at"`
	NextRunAt   time.Time `bson:"next_run_at" json:"next_run_at"`
}
