import { Hex, concatHex } from "viem";

import { findTransaction, findMessageByMessageID } from "../../util/mongodb";
import { config } from "../../util/config";
import { debug, sleep } from "../../util/helpers";
import * as ethereum from "../../util/ethereum";
import * as cosmos from "../../util/cosmos";
import { expect } from "chai";
import { encodeMessage } from "../../util/message";
import { Status } from "../../types";

const isCosmosDomain = (domain: Long) => {
  return cosmos.CHAIN_DOMAIN === domain.toNumber();
};

export const fulfillSignedMessage = async (
  message_id: Hex,
  checkBalance = true,
) => {
  debug("Fulfilling message: ", message_id);
  let message = await findMessageByMessageID(message_id);

  expect(message).to.not.be.null;

  if (!message) return;

  const recipientAddress = message.content.message_body.recipient_address;
  const amount = BigInt(message.content.message_body.amount.toString());

  const ethNetwork = config.ethereum_networks.find(
    (n) => n.chain_id === message?.content.destination_domain.toNumber(),
  );

  expect(ethNetwork).to.not.be.null;

  if (!ethNetwork) return;

  let originChainId = isCosmosDomain(message.content.origin_domain)
    ? config.cosmos_network.chain_id
    : message.content.origin_domain;

  let tx = await findTransaction(
    message.origin_transaction_hash,
    originChainId,
  );

  expect(tx).to.not.be.null;

  if (!tx) return;

  expect(message._id?.toString()).to.oneOf(tx.messages.map((m) => m.toString()));

  expect(message.status).to.be.equal(Status.SIGNED);
  debug("Message signed");

  expect(message.signatures.length).to.be.greaterThanOrEqual(2);

  const beforeWPOKTBalance = await ethereum.getWPOKTBalance(
    ethNetwork.chain_id,
    recipientAddress,
  );

  debug("Fulfilling message...");

  const messageBytes = encodeMessage(message.content);

  const metadata = concatHex(message.signatures.map((s) => s.signature));

  const fulfillmentTx = await ethereum.fulfillOrder(
    ethNetwork.chain_id,
    metadata,
    messageBytes,
  );

  expect(fulfillmentTx).to.not.be.null;

  if (!fulfillmentTx) return;
  debug("Fulfillment Tx: ", fulfillmentTx.transactionHash);

  const fulfillmentEvent = ethereum.findFulfillmentEvent(fulfillmentTx);

  expect(fulfillmentEvent).to.not.be.null;

  if (!fulfillmentEvent) return;

  expect(fulfillmentEvent.orderId.toLowerCase()).to.equal(message.message_id);

  const afterWPOKTBalance = await ethereum.getWPOKTBalance(
    ethNetwork.chain_id,
    recipientAddress,
  );

  if (checkBalance) {
    expect(afterWPOKTBalance).to.equal(beforeWPOKTBalance + amount);
  }
  debug("Fulfillment success");

  await sleep(3000);

  let fulfillmentTxHash = fulfillmentTx.transactionHash.toLowerCase();

  tx = await findTransaction(fulfillmentTxHash, ethNetwork.chain_id);

  expect(tx).to.not.be.null;

  if (!tx) return;
  debug("Fulfillment transaction created");

  expect(tx.status).to.oneOf([Status.PENDING, Status.CONFIRMED]);
  expect(tx.from_address).to.equal(
    await ethereum.getAddress(ethNetwork.chain_id),
  );
  expect(tx.to_address).to.equal(
    ethNetwork.mint_controller_address.toLowerCase(),
  );

  await sleep(4000);

  tx = await findTransaction(fulfillmentTxHash, ethNetwork.chain_id);

  expect(tx).to.not.be.null;

  if (!tx) return;

  expect(tx.status).to.equal(Status.CONFIRMED);
  expect(message._id?.toString()).to.oneOf(tx.messages.map((m) => m.toString()));
  debug("Fulfillment transaction confirmed");

  message = await findMessageByMessageID(message_id);

  expect(message).to.not.be.null;

  if (!message) return;

  expect(message.status).to.equal(Status.SUCCESS);
  expect(message.transaction_hash.toLowerCase()).to.equal(tx.hash);

  expect(message.transaction).to.not.be.null;
  expect(message.transaction?.toString()).to.equal(tx._id?.toString());

  debug("Message success");

  return message;
};
