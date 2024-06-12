// Import the required dependencies for primitive types and date handling
import { ObjectId } from 'mongodb';
import { Hex } from 'viem';

// Assuming Signature type is imported from a local file
import { Signature } from './message';

// Define the RefundStatus type
export type RefundStatus = 'pending' | 'signed' | 'broadcasted' | 'success' | 'invalid';

// Define the Refund type
export type Refund = {
  _id?: ObjectId;
  origin_transaction: ObjectId;
  origin_transaction_hash: Hex;
  recipient: string;
  amount: number;
  transaction_body: string;
  signatures: Signature[];
  transaction?: ObjectId;
  sequence?: number;
  transaction_hash: Hex;
  status: RefundStatus;
  created_at: Date;
  updated_at: Date;
};

export const CollectionRefunds = 'refunds';
