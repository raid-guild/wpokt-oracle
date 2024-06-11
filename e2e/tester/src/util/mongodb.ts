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

const createDatabasePromise = async (): Promise<Db> => {
  const client = new MongoClient(config.mongodb.uri, {});

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

const ensure0xPrefix = (hex: string): string => {
  return hex.startsWith("0x") ? hex.toLowerCase() : `0x${hex.toLowerCase()}`;
}

export const findMessage = async (txHash: string): Promise<Message | null> => {
  const db = await databasePromise;
  return db.collection(CollectionMessages).findOne({
    origin_transaction_hash: ensure0xPrefix(txHash),
  }) as Promise<Message | null>;
};

export const findTransaction = async (
  txHash: string
): Promise<Transaction | null> => {
  const db = await databasePromise;
  return db.collection(CollectionTransactions).findOne({
    hash: ensure0xPrefix(txHash),
  }) as Promise<Transaction | null>;
};

export const findRefund = async (txHash: string): Promise<Refund | null> => {
  const db = await databasePromise;
  return db.collection(CollectionRefunds).findOne({
    origin_transaction_hash: ensure0xPrefix(txHash),
  }) as Promise<Refund | null>;
};
