// Import the required dependencies for primitive types and date handling
import { ObjectId } from 'mongodb';
import { Hex } from 'viem';

// Define the ChainType type
export type ChainType = 'ethereum' | 'cosmos';

// Define the RunnerServiceStatus type
export type RunnerServiceStatus = {
  name: string;
  enabled: boolean;
  blockHeight: number;
  lastRun_at: Date;
  nextRun_at: Date;
};

// Define the Chain type
export type Chain = {
  chain_id: string;
  chain_name: string;
  chain_domain: number;
  chain_type: ChainType;
};

// Define the ChainServiceHealth type
export type ChainServiceHealth = {
  chain: Chain;
  message_monitor?: RunnerServiceStatus;
  message_signer?: RunnerServiceStatus;
  message_relayer?: RunnerServiceStatus;
};

// Define the Node type
export type Node = {
  id?: ObjectId;
  cosmos_address: Hex;
  eth_address: Hex;
  hostname: string;
  oracle_id: string;
  supported_chains: Chain[];
  health: ChainServiceHealth[];
  created_at: Date;
  updated_at: Date;
};

export const CollectionNodes = 'nodes';
