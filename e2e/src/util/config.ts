import yaml from "js-yaml";
import fs from "fs";

const CONFIG_PATH =
  process.env.CONFIG_PATH || "../defaults/config.local.one.yml";

export const HyperlaneVersion = 0; // TODO: This should be changed when contracts are upgraded
export const PausableIsm = "0x5FbDB2315678afecb367f032d93F642f64180aa3";
export const Mailbox = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512";
export const WarpISM = "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9";
export const Token = "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9";
export const MintController = "0x5FC8d32690cc91D4c39d9d3abcBD16989F875707";
export const AccountFactory = "0x610178dA211FEF7D417bC0e6FeD39F05609AD788";
export const Multicall3 = "0xB7f8BC63BbcaD18155201308C8f3540b07f84F5e";

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
  tx_fee: string;
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
