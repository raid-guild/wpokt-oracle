// Import the required dependencies for primitive types and date handling
import { ObjectId } from 'mongodb';
import { Hex } from 'viem';

// Define the MessageContent type
export type MessageContent = {
  version: number;
  nonce: number;
  origin_domain: number;
  sender: Hex;
  destination_domain: number;
  recipient: Hex;
  message_body: MessageBody;
};

// Define the MessageStatus type
export type MessageStatus = 'pending' | 'signed' | 'broadcasted' | 'success' | 'invalid';

// Define the Message type
export type Message = {
  id?: ObjectId;
  origin_transaction: ObjectId;
  origin_transaction_hash: Hex;
  messageId: Hex;
  content: MessageContent;
  transactionBody: Hex;
  signatures: Signature[];
  transaction?: ObjectId;
  sequence?: number;
  transaction_hash: Hex;
  status: MessageStatus;
  created_at: Date;
  updated_at: Date;
};

// Define the MessageBody type
export type MessageBody = {
  sender_address: Hex;
  amount: number;
  recipient_address: Hex;
};

// Define the Signature type
export type Signature = {
  signer: Hex;
  signature: Hex; // Assuming signature is a string representation
};

export const CollectionMessages = 'messages';
