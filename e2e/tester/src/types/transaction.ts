// Import the required dependencies for primitive types and date handling
import { ObjectId } from 'mongodb';
import { Hex } from 'viem';

// Assuming Chain type is imported from a local file
import { Chain } from './node';

// Define the TransactionStatus type
export type TransactionStatus = 'pending' | 'confirmed' | 'failed' | 'invalid';

// Define the Transaction type
export type Transaction = {
  id?: ObjectId;
  hash: Hex;
  from_address: Hex;
  to_address: Hex;
  block_height: number;
  confirmations: number;
  chain: Chain;
  status: TransactionStatus;
  created_at: Date;
  updated_at: Date;
  refund?: ObjectId;
  messages: ObjectId[];
};

// Define the MintMemo type
export type MintMemo = {
  address: Hex;
  chain_id: string;
};

export const CollectionTransactions = 'transactions';
