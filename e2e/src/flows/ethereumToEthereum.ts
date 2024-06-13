import { Hex, concatHex, parseUnits } from "viem";
import * as ethereum from "../util/ethereum";
import { expect } from "chai";
import { config, HYPERLANE_VERSION } from "../util/config";
import { Status } from "../types";
import { findMessage, findTransaction } from "../util/mongodb";
import { sleep, debug } from "../util/helpers";
import { addressHexToBytes32, decodeMessage, encodeMessage } from "../util/message";


export const ethereumToEthereumFlow = async () => {
  const ethNetworkOne = config.ethereum_networks[0];
  const ethNetworkTwo = config.ethereum_networks[1];

  it("should initiate order and fulfill order on the other chain", async () => {
    debug("\nTesting -- should initiate order and fulfill order on the other chain");

    const fromAddress = await ethereum.getAddress(ethNetworkOne.chain_id);
    const recipientAddress = fromAddress;
    const destMintControllerAddress = ethNetworkTwo.mint_controller_address.toLowerCase() as Hex;
    const amount = parseUnits("1", 6);

    const fromBeforeBalance = await ethereum.getWPOKTBalance(ethNetworkOne.chain_id, fromAddress);
    const beforeWPOKTBalance = await ethereum.getWPOKTBalance(ethNetworkTwo.chain_id, recipientAddress);

    expect(Number(fromBeforeBalance)).to.be.greaterThan(Number(amount));

    debug("Sending transaction...");
    const dispatchTx = await ethereum.initiateOrder(
      ethNetworkOne.chain_id,
      ethNetworkTwo.chain_id,
      recipientAddress,
      amount
    );

    expect(dispatchTx).to.not.be.null;

    if (!dispatchTx) return;
    debug("Transaction sent: ", dispatchTx.transactionHash);

    const dispatchEvent = ethereum.findDispatchEvent(dispatchTx);

    expect(dispatchEvent).to.not.be.null;

    if (!dispatchEvent) return;

    debug("Dispatch event found");

    expect(dispatchEvent.recipient.toLowerCase()).to.equal(addressHexToBytes32(destMintControllerAddress));
    expect(dispatchEvent.sender.toLowerCase()).to.equal(ethNetworkOne.mint_controller_address.toLowerCase());
    expect(dispatchEvent.destination).to.equal(ethNetworkTwo.chain_id);

    const messageContent = decodeMessage(dispatchEvent.message);

    expect(messageContent).to.not.be.null;

    if (!messageContent) return;

    debug("Message found");

    expect(messageContent.version).to.equal(HYPERLANE_VERSION);
    // expect(messageContent.nonce).to.equal(0);
    expect(messageContent.origin_domain).to.equal(ethNetworkOne.chain_id);
    expect(messageContent.sender).to.equal(ethNetworkOne.mint_controller_address.toLowerCase());
    expect(messageContent.destination_domain).to.equal(ethNetworkTwo.chain_id);
    expect(messageContent.recipient).to.equal(destMintControllerAddress);
    expect(messageContent.message_body.sender_address).to.equal(fromAddress);
    expect(messageContent.message_body.recipient_address).to.equal(recipientAddress);
    expect(messageContent.message_body.amount.toString()).to.equal(amount.toString());


    debug("Waiting for transaction to be created...");
    await sleep(2000);

    const originTxHash = dispatchTx.transactionHash.toLowerCase();

    let tx = await findTransaction(originTxHash);

    expect(tx).to.not.be.null;

    if (!tx) return;

    debug("Transaction created");

    expect(tx.hash).to.equal(originTxHash);
    expect(tx.from_address).to.equal(fromAddress);
    expect(tx.to_address).to.equal(ethNetworkOne.mailbox_address.toLowerCase());
    expect(tx.block_height.toString()).to.equal(dispatchTx.blockNumber.toString());
    expect(tx.status).to.be.oneOf([Status.PENDING, Status.CONFIRMED]);

    debug("Waiting for transaction to be confirmed...");
    await sleep(5000);

    tx = await findTransaction(originTxHash);

    expect(tx).to.not.be.null;

    if (!tx) return;

    expect(tx.status).to.be.equal(Status.CONFIRMED);

    debug("Transaction confirmed");

    let message = await findMessage(originTxHash);

    expect(message).to.not.be.null;

    if (!message) return;

    debug("Message created");

    expect(message.origin_transaction_hash).to.equal(originTxHash);
    expect(message.origin_transaction?.toString()).to.equal(tx._id?.toString());
    expect(message.content).to.deep.equal(messageContent);
    expect(message.status).to.be.oneOf([Status.PENDING, Status.SIGNED]);

    await sleep(3500);

    message = await findMessage(originTxHash);

    expect(message).to.not.be.null;

    if (!message) return;

    expect(message.status).to.be.equal(Status.SIGNED);

    debug("Message signed");

    await sleep(2000);


    tx = await findTransaction(originTxHash);

    expect(tx).to.not.be.null;

    if (!tx) return;

    expect(tx.messages.length).to.equal(1);
    expect(tx.messages[0].toString()).to.equal(message._id?.toString());

    expect(message.signatures.length).to.be.greaterThanOrEqual(2);


    debug("Fulfilling Order...");

    const messageBytes = encodeMessage(message.content);

    const metadata = concatHex(message.signatures.map((s) => s.signature));

    const fulfillmentTx = await ethereum.fulfillOrder(ethNetworkTwo.chain_id, metadata, messageBytes);

    expect(fulfillmentTx).to.not.be.null;

    if (!fulfillmentTx) return;
    debug("Fulfilled: ", fulfillmentTx.transactionHash);

    const fulfillmentEvent = ethereum.findFulfillmentEvent(fulfillmentTx);

    expect(fulfillmentEvent).to.not.be.null;

    if (!fulfillmentEvent) return;

    expect(fulfillmentEvent.orderId.toLowerCase()).to.equal(message.message_id);

    debug("Fulfillment success");

    await sleep(3000);

    const txHash = fulfillmentTx.transactionHash.toLowerCase();

    tx = await findTransaction(txHash);

    expect(tx).to.not.be.null;

    if (!tx) return;
    debug("Fulfillment transaction created");

    expect(tx.status).to.oneOf([Status.PENDING, Status.CONFIRMED]);
    expect(tx.from_address).to.equal(await ethereum.getAddress(ethNetworkTwo.chain_id));
    expect(tx.to_address).to.equal(ethNetworkTwo.mint_controller_address.toLowerCase());

    await sleep(3000);

    tx = await findTransaction(txHash);

    expect(tx).to.not.be.null;

    if (!tx) return;

    expect(tx.status).to.equal(Status.CONFIRMED);
    expect(tx.messages.length).to.equal(1);
    expect(tx.messages[0].toString()).to.equal(message._id?.toString());

    message = await findMessage(originTxHash);

    expect(message).to.not.be.null;

    if (!message) return;

    expect(message.status).to.equal(Status.SUCCESS);
    expect(message.transaction_hash.toLowerCase()).to.equal(tx.hash);

    expect(message.transaction).to.not.be.null;
    expect(message.transaction?.toString()).to.equal(tx._id?.toString());

    debug("Fullfillment success");

    const afterWPOKTBalance = await ethereum.getWPOKTBalance(ethNetworkTwo.chain_id, recipientAddress);

    expect(afterWPOKTBalance).to.equal(beforeWPOKTBalance + amount);

    const fromAfterBalance = await ethereum.getWPOKTBalance(ethNetworkOne.chain_id, fromAddress);
    expect(fromAfterBalance).to.equal(fromBeforeBalance - amount);
  });
};
