import yaml from "js-yaml";
import fs from "fs";

const CONFIG_PATH =
  process.env.CONFIG_PATH || "../config/tester/config.local.yml";

// TODO: This should be changed 
export const HYPERLANE_VERSION = 0;

export type Config = {
  mnemonic: string;
  health_check: HealthCheckConfig;
  logger: LoggerConfig;
  mongodb: MongoConfig;
  ethereum_networks: EthereumNetworkConfig[];
  cosmos_network: CosmosNetworkConfig;
};

export type HealthCheckConfig = {
  interval_ms: number;
  read_last_health: boolean;
};

export type LoggerConfig = {
  level: string;
  format: string; // json or text
};

export type MongoConfig = {
  uri: string;
  database: string;
  timeout_ms: number;
};

export type EthereumNetworkConfig = {
  start_block_height: number;
  confirmations: number;
  rpc_url: string;
  timeout_ms: number;
  chain_id: number;
  chain_name: string;
  mailbox_address: string;
  mint_controller_address: string;
  omni_token_address: string;
  warp_ism_address: string;
  oracle_addresses: string[];
  message_monitor: ServiceConfig;
  message_signer: ServiceConfig;
  message_relayer: ServiceConfig;
};

export type CosmosNetworkConfig = {
  start_block_height: number;
  confirmations: number;
  rpc_url: string;
  grpc_enabled: boolean;
  grpc_host: string;
  grpc_port: number;
  timeout_ms: number;
  chain_id: string;
  chain_name: string;
  tx_fee: number;
  bech32_prefix: string;
  coin_denom: string;
  multisig_address: string;
  multisig_public_keys: string[];
  multisig_threshold: number;
  message_monitor: ServiceConfig;
  message_signer: ServiceConfig;
  message_relayer: ServiceConfig;
};

export type ServiceConfig = {
  enabled: boolean;
  interval_ms: number;
};


export const config = yaml.load(fs.readFileSync(CONFIG_PATH, "utf8")) as Config;
