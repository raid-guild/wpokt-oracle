import yaml from "js-yaml";
import fs from "fs";

const CONFIG_PATH =
  process.env.CONFIG_PATH || "../defaults/config.local.one.yml";

export const HyperlaneVersion = 3;
export const PausableIsm = "0x5FbDB2315678afecb367f032d93F642f64180aa3";
export const Mailbox = "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9";
export const WarpISM = "0x0165878A594ca255338adfa4d48449f69242Eb8F";
export const Token = "0xa513E6E4b8f2a923D98304ec87F64353C4D5C853";
export const MintController = "0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6";
export const AccountFactory = "0x0DCd1Bf9A1b36cE34237eEaFef220932846BCD82";
export const Multicall3 = "0x9A676e781A523b5d0C0e43731313A708CB607508";


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
