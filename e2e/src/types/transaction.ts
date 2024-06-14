// Import the required dependencies for primitive types and date handling
import { ObjectId, Long } from 'mongodb';
import { Hex } from 'viem';

// Assuming Chain type is imported from a local file
import { Chain } from './node';

// Define the TransactionStatus type
export type TransactionStatus = 'pending' | 'confirmed' | 'failed' | 'invalid';

// Define the Transaction type
export type Transaction = {
  readonly _id: ObjectId;
  readonly hash: Hex;
  readonly from_address: Hex;
  readonly to_address: Hex;
  readonly block_height: Long;
  readonly confirmations: Long;
  readonly chain: Chain;
  readonly status: TransactionStatus;
  readonly created_at: Date;
  readonly updated_at: Date;
  readonly refund?: ObjectId | null;
  readonly messages: ObjectId[];
};

// Define the MintMemo type
export type MintMemo = {
  readonly address: Hex;
  readonly chain_id: string;
};

export const CollectionTransactions = 'transactions';
