import { parseUnits } from "viem";
import * as ethereum from "../util/ethereum";
import * as cosmos from "../util/cosmos";
import { expect } from "chai";
import { config, HYPERLANE_VERSION } from "../util/config";
import { Message, Status } from "../types";
import { findMessageByMessageID, findMessagesByTxHash, findTransaction } from "../util/mongodb";
import { sleep, debug } from "../util/helpers";
import { addressHexToBytes32, decodeMessage } from "../util/message";

const POKT_TX_FEE = BigInt(0);

export const ethereumToCosmosFlow = async () => {
  const ethNetwork = config.ethereum_networks[0];
  const cosmosNetwork = config.cosmos_network;

  it("should initiate order and transfer amount from vault", async () => {
    debug("\nTesting -- should initiate order and transfer amount from vault");

    const fromAddress = await ethereum.getAddress(ethNetwork.chain_id);
    const recipientBech32 = await cosmos.getAddress();
    const recipientAddress = cosmos.bech32ToHex(recipientBech32);
    const destMintControllerAddress = cosmos.bech32ToHex(cosmosNetwork.multisig_address);
    const amount = parseUnits("1", 6);

    const recipientBeforeBalance = await cosmos.getBalance(recipientBech32);
    const fromBeforeBalance = await ethereum.getWPOKTBalance(ethNetwork.chain_id, fromAddress);

    expect(Number(fromBeforeBalance)).to.be.greaterThan(Number(amount));

    debug("Sending transaction...");
    const dispatchTx = await ethereum.initiateOrder(
      ethNetwork.chain_id,
      cosmos.CHAIN_DOMAIN,
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
    expect(dispatchEvent.sender.toLowerCase()).to.equal(ethNetwork.mint_controller_address.toLowerCase());
    expect(dispatchEvent.destination).to.equal(cosmos.CHAIN_DOMAIN);

    const messageContent = decodeMessage(dispatchEvent.message);

    expect(messageContent).to.not.be.null;

    if (!messageContent) return;

    debug("Message found");

    expect(messageContent.version).to.equal(HYPERLANE_VERSION);
    // expect(messageContent.nonce).to.equal(0);
    expect(messageContent.origin_domain.toNumber()).to.equal(ethNetwork.chain_id);
    expect(messageContent.sender).to.equal(ethNetwork.mint_controller_address.toLowerCase());
    expect(messageContent.destination_domain.toNumber()).to.equal(cosmos.CHAIN_DOMAIN);
    expect(messageContent.recipient).to.equal(destMintControllerAddress);
    expect(messageContent.message_body.sender_address).to.equal(fromAddress);
    expect(messageContent.message_body.recipient_address).to.equal(recipientAddress);
    expect(messageContent.message_body.amount.toString()).to.equal(amount.toString());


    debug("Waiting for transaction to be created...");
    await sleep(2000);

    const originTxHash = dispatchTx.transactionHash.toLowerCase();

    let tx = await findTransaction(originTxHash, ethNetwork.chain_id);

    expect(tx).to.not.be.null;

    if (!tx) return;

    debug("Transaction created");

    expect(tx.hash).to.equal(originTxHash);
    expect(tx.from_address).to.equal(fromAddress);
    expect(tx.to_address).to.equal(ethNetwork.mailbox_address.toLowerCase());
    expect(tx.block_height.toString()).to.equal(dispatchTx.blockNumber.toString());
    expect(tx.status).to.be.oneOf([Status.PENDING, Status.CONFIRMED]);

    debug("Waiting for message to be confirmed...");
    await sleep(5000);

    tx = await findTransaction(originTxHash, ethNetwork.chain_id);

    expect(tx).to.not.be.null;

    if (!tx) return;

    expect(tx.status).to.be.equal(Status.CONFIRMED);

    debug("Transaction confirmed");

    let messages = await findMessagesByTxHash(originTxHash);

    expect(messages).to.not.be.null;
    expect(messages.length).to.equal(1);

    let message: Message | null = messages[0];

    expect(message).to.not.be.null;

    if (!message) return;

    debug("Message created");

    expect(message.origin_transaction_hash).to.equal(originTxHash);
    expect(message.origin_transaction?.toString()).to.equal(tx._id?.toString());
    expect(message.content).to.deep.equal(messageContent);
    expect(message.status).to.be.oneOf([Status.PENDING, Status.SIGNED, Status.BROADCASTED, Status.SUCCESS]);

    await sleep(3500);

    message = await findMessageByMessageID(message.message_id);

    expect(message).to.not.be.null;

    if (!message) return;

    expect(message.status).to.be.oneOf([Status.SIGNED, Status.BROADCASTED, Status.SUCCESS]);

    debug("Message signed");

    await sleep(2000);

    message = await findMessageByMessageID(message.message_id);

    expect(message).to.not.be.null;

    if (!message) return;

    expect(message.status).to.be.oneOf([Status.BROADCASTED, Status.SUCCESS]);
    expect(message.transaction_hash).to.not.be.null;


    const txHash = message.transaction_hash.toLowerCase();

    debug("Message broadcasted");

    tx = await findTransaction(txHash, config.cosmos_network.chain_id);

    expect(tx).to.not.be.null;

    if (!tx) return;

    debug("Message transaction created");

    expect(tx.status).to.oneOf([Status.PENDING, Status.CONFIRMED]);
    expect(tx.from_address).to.equal(destMintControllerAddress);
    expect(tx.to_address).to.equal(recipientAddress);

    await sleep(3500);

    tx = await findTransaction(txHash, config.cosmos_network.chain_id);

    expect(tx).to.not.be.null;

    if (!tx) return;

    expect(tx.status).to.equal(Status.CONFIRMED);
    expect(tx.messages.length).to.equal(1);
    expect(tx.messages[0].toString()).to.equal(message._id?.toString());

    message = await findMessageByMessageID(message.message_id);

    expect(message).to.not.be.null;

    if (!message) return;

    expect(message.status).to.equal(Status.SUCCESS);
    expect(message.transaction_hash.toLowerCase()).to.equal(tx.hash);

    expect(message.transaction).to.not.be.null;
    expect(message.transaction?.toString()).to.equal(tx._id?.toString());

    const recipientAfterBalance = await cosmos.getBalance(recipientBech32);
    expect(recipientAfterBalance).to.equal(
      recipientBeforeBalance + amount - POKT_TX_FEE
    );

    const fromAfterBalance = await ethereum.getWPOKTBalance(ethNetwork.chain_id, fromAddress);
    expect(fromAfterBalance).to.equal(fromBeforeBalance - amount);

    debug("Message success");

  });
};
