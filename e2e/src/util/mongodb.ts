import { Db, MongoClient } from "mongodb";

import { config } from "./config";
import {
  Node,
  CollectionNodes,
  Transaction,
  CollectionTransactions,
  Refund,
  CollectionRefunds,
  Message,
  CollectionMessages,
} from "../types";
import { Hex } from "viem";

const createDatabasePromise = async (): Promise<Db> => {
  const client = new MongoClient(config.mongodb.uri, { promoteLongs: false });

  await client.connect();
  return client.db(config.mongodb.database);
};

export const databasePromise: Promise<Db> = createDatabasePromise();

export const findNodes = async (): Promise<Node[]> => {
  const db = await databasePromise;
  return db.collection(CollectionNodes)
    .find({})
    .toArray() as unknown as Promise<Node[]>;
};

const ensure0xPrefix = (hex: string): Hex => {
  return hex.startsWith("0x") ? hex.toLowerCase() as Hex : `0x${hex.toLowerCase()}`;
}

export const findMessagesByTxHash = async (txHash: string): Promise<Message[]> => {
  const db = await databasePromise;
  return db.collection(CollectionMessages).find({
    origin_transaction_hash: ensure0xPrefix(txHash),
  }).toArray() as unknown as Promise<Message[]>;
};

export const findMessageByMessageID = async (messageID: Hex): Promise<Message | null> => {
  const db = await databasePromise;
  return db.collection(CollectionMessages).findOne({
    message_id: ensure0xPrefix(messageID),
  }) as Promise<Message | null>;
}

export const findTransaction = async (
  txHash: string,
  chain_id: string | number | bigint | Long,
): Promise<Transaction | null> => {
  const db = await databasePromise;
  return db.collection(CollectionTransactions).findOne({
    hash: ensure0xPrefix(txHash),
    "chain.chain_id": chain_id.toString(),
  }) as Promise<Transaction | null>;
};

export const findRefund = async (txHash: string): Promise<Refund | null> => {
  const db = await databasePromise;
  return db.collection(CollectionRefunds).findOne({
    origin_transaction_hash: ensure0xPrefix(txHash),
  }) as Promise<Refund | null>;
};
