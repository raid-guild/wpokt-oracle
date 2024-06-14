// Import the required dependencies for primitive types and date handling
import { ObjectId, Long } from "mongodb";
import { Hex } from "viem";

// Assuming Signature type is imported from a local file
import { Signature } from "./message";

// Define the RefundStatus type
export type RefundStatus =
  | "pending"
  | "signed"
  | "broadcasted"
  | "success"
  | "invalid";

// Define the Refund type
export type Refund = {
  readonly _id: ObjectId;
  readonly origin_transaction: ObjectId;
  readonly origin_transaction_hash: Hex;
  readonly recipient: string;
  readonly amount: Long;
  readonly transaction_body: string;
  readonly signatures: Signature[];
  readonly transaction?: ObjectId | null;
  readonly sequence?: Long | null;
  readonly transaction_hash: Hex;
  readonly status: RefundStatus;
  readonly created_at: Date;
  readonly updated_at: Date;
};

export const CollectionRefunds = "refunds";
