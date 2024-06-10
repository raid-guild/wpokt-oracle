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

export const findMessage = async (txHash: string): Promise<Message | null> => {
  const db = await databasePromise;
  return db.collection(CollectionMessages).findOne({
    origin_transaction_hash: txHash.toLowerCase(),
  }) as Promise<Message | null>;
};

export const findTransaction = async (
  txHash: string
): Promise<Transaction | null> => {
  const db = await databasePromise;
  return db.collection(CollectionTransactions).findOne({
    hash: txHash.toLowerCase(),
  }) as Promise<Transaction | null>;
};

export const findRefund = async (txHash: string): Promise<Refund | null> => {
  const db = await databasePromise;
  return db.collection(CollectionRefunds).findOne({
    origin_transaction_hash: txHash.toLowerCase(),
  }) as Promise<Refund | null>;
};
