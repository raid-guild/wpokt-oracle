import yaml from "js-yaml";
import fs from "fs";

const CONFIG_PATH =
  process.env.CONFIG_PATH || "../defaults/config.local.yml";

export const HyperlaneVersion = 3;
export const Mailbox = "0x9fE46736679d2D9a65F0992F2272dE9f3c7fa6e0";
export const WarpISM = "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9";
export const Token = "0x5FC8d32690cc91D4c39d9d3abcBD16989F875707";
export const MintController = "0x0165878A594ca255338adfa4d48449f69242Eb8F";
export const AccountFactory = "0xB7f8BC63BbcaD18155201308C8f3540b07f84F5e";
export const Multicall3 = "0xA51c1fc2f0D1a1b8494Ed1FE312d7C3a78Ed91C0";

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
