package models

type Config struct {
	Mnemonic         string                  `yaml:"mnemonic" json:"mnemonic"`
	HealthCheck      HealthCheckConfig       `yaml:"health_check" json:"health_check"`
	Logger           LoggerConfig            `yaml:"logger" json:"logger"`
	MongoDB          MongoConfig             `yaml:"mongodb" json:"mongodb"`
	EthereumNetworks []EthereumNetworkConfig `yaml:"ethereum_networks" json:"ethereum_networks"`
	CosmosNetworks   []CosmosNetworkConfig   `yaml:"cosmos_networks" json:"cosmos_networks"`
}

type HealthCheckConfig struct {
	IntervalMS     uint64 `yaml:"interval_ms" json:"interval_ms"`
	ReadLastHealth bool   `yaml:"read_last_health" json:"read_last_health"`
}

type LoggerConfig struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"` // json or text
}

type MongoConfig struct {
	URI       string `yaml:"uri" json:"uri"`
	Database  string `yaml:"database" json:"database"`
	TimeoutMS uint64 `yaml:"timeout_ms" json:"timeout_ms"`
}

type EthereumNetworkConfig struct {
	StartBlockHeight      uint64        `yaml:"start_block_height" json:"start_block_height"`
	Confirmations         uint64        `yaml:"confirmations" json:"confirmations"`
	RPCURL                string        `yaml:"rpc_url" json:"rpcurl"`
	TimeoutMS             uint64        `yaml:"timeout_ms" json:"timeout_ms"`
	ChainID               uint64        `yaml:"chain_id" json:"chain_id"`
	ChainName             string        `yaml:"chain_name" json:"chain_name"`
	MailboxAddress        string        `yaml:"mailbox_address" json:"mailbox_address"`
	MintControllerAddress string        `yaml:"mint_controller_address" json:"mint_controller_address"`
	OracleAddresses       []string      `yaml:"oracle_addresses" json:"oracle_addresses"`
	MessageMonitor        ServiceConfig `yaml:"message_monitor" json:"message_monitor"`
	MessageSigner         ServiceConfig `yaml:"message_signer" json:"message_signer"`
	MessageRelayer        ServiceConfig `yaml:"message_relayer" json:"message_relayer"`
}

type CosmosNetworkConfig struct {
	StartBlockHeight   uint64        `yaml:"start_block_height" json:"start_block_height"`
	Confirmations      uint64        `yaml:"confirmations" json:"confirmations"`
	RPCURL             string        `yaml:"rpc_url" json:"rpcurl"`
	GRPCEnabled        bool          `yaml:"grpc_enabled" json:"grpc_enabled"`
	GRPCHost           string        `yaml:"grpc_host" json:"grpc_host"`
	GRPCPort           uint64        `yaml:"grpc_port" json:"grpc_port"`
	TimeoutMS          uint64        `yaml:"timeout_ms" json:"time_out_ms"`
	ChainID            string        `yaml:"chain_id" json:"chain_id"`
	ChainName          string        `yaml:"chain_name" json:"chain_name"`
	TxFee              uint64        `yaml:"tx_fee" json:"tx_fee"`
	Bech32Prefix       string        `yaml:"bech32_prefix" json:"bech32_prefix"`
	CoinDenom          string        `yaml:"coin_denom" json:"coin_denom"`
	MultisigAddress    string        `yaml:"multisig_address" json:"multisig_address"`
	MultisigPublicKeys []string      `yaml:"multisig_public_keys" json:"multisig_public_keys"`
	MultisigThreshold  uint64        `yaml:"multisig_threshold" json:"multisig_threshold"`
	MessageMonitor     ServiceConfig `yaml:"message_monitor" json:"message_monitor"`
	MessageSigner      ServiceConfig `yaml:"message_signer" json:"message_signer"`
	MessageRelayer     ServiceConfig `yaml:"message_relayer" json:"message_relayer"`
}

type ServiceConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	IntervalMS uint64 `yaml:"interval_ms" json:"interval_ms"`
}
