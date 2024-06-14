// Import the required dependencies for primitive types and date handling
import { ObjectId, Long } from 'mongodb';
import { Hex } from 'viem';

// Define the ChainType type
export type ChainType = 'ethereum' | 'cosmos';

// Define the RunnerServiceStatus type
export type RunnerServiceStatus = {
  readonly name: string;
  readonly enabled: boolean;
  readonly blockHeight: Long;
  readonly lastRun_at: Date;
  readonly nextRun_at: Date;
};

// Define the Chain type
export type Chain = {
  readonly chain_id: string;
  readonly chain_name: string;
  readonly chain_domain: Long;
  readonly chain_type: ChainType;
};

// Define the ChainServiceHealth type
export type ChainServiceHealth = {
  readonly chain: Chain;
  readonly message_monitor?: RunnerServiceStatus | null;
  readonly message_signer?: RunnerServiceStatus | null;
  readonly message_relayer?: RunnerServiceStatus | null;
};

// Define the Node type
export type Node = {
  readonly _id: ObjectId;
  readonly cosmos_address: Hex;
  readonly eth_address: Hex;
  readonly hostname: string;
  readonly oracle_id: string;
  readonly supported_chains: Chain[];
  readonly health: ChainServiceHealth[];
  readonly created_at: Date;
  readonly updated_at: Date;
};

export const CollectionNodes = 'nodes';
