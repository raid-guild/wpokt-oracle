import yaml from "js-yaml";
import fs from "fs";
import { Hex } from "viem";

const CONFIG_PATH =
  process.env.CONFIG_PATH || "../defaults/config.local.yml";


export type Config = {
  logger: LoggerConfig;
  mongodb: MongoConfig;
  ethereum_networks: EthereumNetworkConfig[];
  cosmos_network: CosmosNetworkConfig;
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
};

export const config = yaml.load(fs.readFileSync(CONFIG_PATH, "utf8")) as Config;

export const HyperlaneVersion = 3;
export const Mailbox = config.ethereum_networks[0].mailbox_address as Hex;
export const WarpISM = config.ethereum_networks[0].warp_ism_address as Hex;
export const Token = config.ethereum_networks[0].omni_token_address as Hex;
export const MintController = config.ethereum_networks[0].mint_controller_address as Hex;
export const AccountFactory = "0xB7f8BC63BbcaD18155201308C8f3540b07f84F5e";
export const Multicall3 = "0xA51c1fc2f0D1a1b8494Ed1FE312d7C3a78Ed91C0";
