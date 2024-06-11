import { TransactionReceipt, concatHex, parseUnits } from "viem";
import * as ethereum from "../util/ethereum";
import * as cosmos from "../util/cosmos";
import { expect } from "chai";
import { sleep, debug } from "../util/helpers";
import { config } from "../util/config";
import { Message, MintMemo, Status } from "../types";
import { findMessage, findRefund, findTransaction } from "../util/mongodb";
import { encodeMessage } from "../util/message";

const POKT_TX_FEE = BigInt(0);

export const cosmosToEthereumFlow = async () => {
  const ethNetwork = config.ethereum_networks[0];
  const cosmosNetwork = config.cosmos_network;

  it("should fulfill on ethereum for send tx to cosmos vault with valid memo", async () => {

    debug("\nTesting -- should fulfill on ethereum for send tx to cosmos vault with valid memo");

    const signer = await cosmos.signerPromise;
    const fromAddress = await cosmos.getAddress();
    const recipientAddress = await ethereum.getAddress(ethNetwork.chain_id);
    const toAddress = cosmosNetwork.multisig_address;
    const amount = parseUnits("1", 6);

    const memo: MintMemo = {
      address: recipientAddress,
      chain_id: ethNetwork.chain_id.toString(),
    };

    const fromBeforeBalance = await cosmos.getBalance(fromAddress);
    const toBeforeBalance = await cosmos.getBalance(toAddress);

    debug("Sending transaction...");
    const sendTx = await cosmos.sendPOKT(
      signer,
      toAddress,
      amount.toString(),
      JSON.stringify(memo),
      POKT_TX_FEE.toString(),
    );

    expect(sendTx).to.not.be.null;

    if (!sendTx) return;
    debug("Transaction sent: ", sendTx.hash);

    expect(sendTx.hash).to.be.a("string");
    expect(sendTx.hash).to.have.lengthOf(64);
    expect(sendTx.code).to.equal(0);

    const fromAfterBalance = await cosmos.getBalance(fromAddress);
    const toAfterBalance = await cosmos.getBalance(toAddress);

    expect(fromAfterBalance).to.equal(fromBeforeBalance - amount - POKT_TX_FEE);
    expect(toAfterBalance).to.equal(toBeforeBalance + amount);

    debug("Waiting for transaction to be created...");
    await sleep(1500);

    let txHash = "0x" + sendTx.hash.toLowerCase();
    const originTxHash = txHash;

    let tx = await findTransaction(txHash);

    expect(tx).to.not.be.null;

    if (!tx) return;
    debug("Transaction created");

    const fromHex = cosmos.bech32ToHex(fromAddress);
    const toHex = cosmos.bech32ToHex(toAddress);

    expect(tx.block_height.toString()).to.equal(sendTx.height.toString());
    expect(tx.from_address).to.equal(fromHex);
    expect(tx.to_address).to.equal(toHex);
    expect(tx.status).to.oneOf([Status.PENDING, Status.CONFIRMED]);

    await sleep(3000);

    tx = await findTransaction(txHash);

    expect(tx).to.not.be.null;

    if (!tx) return;

    expect(tx.status).to.equal(Status.CONFIRMED);
    debug("Transaction confirmed");

    let message = await findMessage(txHash);

    expect(message).to.not.be.null;

    if (!message) return;
    debug("Message created");

    expect(message.origin_transaction_hash).to.equal(txHash);
    expect(message.origin_transaction.toString()).to.equal(tx._id?.toString());

    const account = await cosmos.getAccount(fromAddress);

    expect(account).to.not.be.null;

    if (!account) return;

    expect(message.content.version).to.equal(cosmos.HYPERLANE_VERSION);
    expect(message.content.nonce).to.equal(account.sequence - 1);
    expect(message.content.origin_domain).to.equal(cosmos.CHAIN_DOMAIN);
    expect(message.content.sender).to.equal(fromHex);
    expect(message.content.destination_domain).to.equal(ethNetwork.chain_id);
    expect(message.content.recipient).to.equal(ethNetwork.mint_controller_address.toLowerCase());
    expect(message.content.message_body.sender_address).to.equal(fromHex);
    expect(message.content.message_body.recipient_address).to.equal(recipientAddress.toLowerCase());
    expect(message.content.message_body.amount.toString()).to.equal(amount.toString());

    expect(message.status).to.oneOf([Status.PENDING, Status.SIGNED]);

    await sleep(2200);

    tx = await findTransaction(txHash);

    expect(tx).to.not.be.null;

    if (!tx) return;

    expect(tx.messages.length).to.equal(1);
    expect(tx.messages[0].toString()).to.equal(message._id?.toString());

    message = await findMessage(txHash);

    expect(message).to.not.be.null;

    if (!message) return;

    expect(message.status).to.be.equal(Status.SIGNED);
    debug("Message signed");

    expect(message.signatures.length).to.be.greaterThanOrEqual(2);

    const beforeWPOKTBalance = await ethereum.getWPOKTBalance(ethNetwork.chain_id, recipientAddress);

    debug("Fulfilling Order...");

    const messageBytes = encodeMessage(message.content);

    const metadata = concatHex(message.signatures.map((s) => s.signature));

    const fulfillmentTx = await ethereum.fulfillOrder(ethNetwork.chain_id, metadata, messageBytes);

    expect(fulfillmentTx).to.not.be.null;

    if (!fulfillmentTx) return;
    debug("Fulfilled: ", fulfillmentTx.transactionHash);

    const fulfillmentEvent = ethereum.findFulfillmentEvent(fulfillmentTx);

    expect(fulfillmentEvent).to.not.be.null;

    if (!fulfillmentEvent) return;

    expect(fulfillmentEvent.orderId.toLowerCase()).to.equal(message.message_id);

    const afterWPOKTBalance = await ethereum.getWPOKTBalance(ethNetwork.chain_id, recipientAddress);

    expect(afterWPOKTBalance).to.equal(beforeWPOKTBalance + amount);
    debug("Fulfillment success");

    await sleep(3000);

    txHash = fulfillmentTx.transactionHash.toLowerCase();

    tx = await findTransaction(txHash);

    expect(tx).to.not.be.null;

    if (!tx) return;
    debug("Fulfillment transaction created");

    expect(tx.status).to.oneOf([Status.PENDING, Status.CONFIRMED]);
    expect(tx.from_address).to.equal(await ethereum.getAddress(ethNetwork.chain_id));
    expect(tx.to_address).to.equal(ethNetwork.mint_controller_address.toLowerCase());

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
  });


  it("should refund amount for send tx to vault with invalid memo", async () => {

    debug("\nTesting -- should refund amount for send tx to vault with invalid memo");

    const signer = await cosmos.signerPromise;
    const fromAddress = await cosmos.getAddress();
    const toAddress = cosmosNetwork.multisig_address;
    const amount = parseUnits("1", 6);

    const beforeBalance = await cosmos.getBalance(fromAddress);

    const memo = "not a json";

    debug("Sending transaction...");
    const sendTx = await cosmos.sendPOKT(
      signer,
      toAddress,
      amount.toString(),
      memo,
      POKT_TX_FEE.toString(),
    );

    expect(sendTx).to.not.be.null;

    if (!sendTx) return;
    debug("Transaction sent: ", sendTx.hash);

    expect(sendTx.hash).to.be.a("string");
    expect(sendTx.hash).to.have.lengthOf(64);
    expect(sendTx.code).to.equal(0);

    debug("Waiting for transaction to be created...");
    await sleep(1500);

    let txHash = "0x" + sendTx.hash.toLowerCase();
    const originTxHash = txHash;

    let tx = await findTransaction(txHash);

    expect(tx).to.not.be.null;

    if (!tx) return;
    debug("Transaction created");

    const fromHex = cosmos.bech32ToHex(fromAddress);
    const toHex = cosmos.bech32ToHex(toAddress);

    expect(tx.block_height.toString()).to.equal(sendTx.height.toString());
    expect(tx.from_address).to.equal(fromHex);
    expect(tx.to_address).to.equal(toHex);
    expect(tx.status).to.oneOf([Status.PENDING, Status.CONFIRMED]);

    const account = await cosmos.getAccount(cosmosNetwork.multisig_address);

    expect(account).to.not.be.null;

    if (!account) return;

    await sleep(2500);

    tx = await findTransaction(txHash);

    expect(tx).to.not.be.null;

    if (!tx) return;

    expect(tx.status).to.equal(Status.CONFIRMED);
    debug("Transaction confirmed");

    let refund = await findRefund(txHash);

    expect(refund).to.not.be.null;

    if (!refund) return;
    debug("Refund created");

    expect(refund.origin_transaction_hash).to.equal(txHash);
    expect(refund.origin_transaction.toString()).to.equal(tx._id?.toString());

    expect(refund.recipient).to.equal(fromHex);
    expect(refund.amount.toString()).to.equal(amount.toString());
    expect(refund.transaction_body).to.not.be.null;
    expect(refund.transaction_body).to.not.equal("");

    expect(refund.status).to.oneOf([Status.PENDING, Status.SIGNED, Status.BROADCASTED]);

    await sleep(2200);

    tx = await findTransaction(txHash);

    expect(tx).to.not.be.null;

    if (!tx) return;

    expect(tx.refund).to.not.be.null;
    if (!tx.refund) return;
    expect(tx.refund.toString()).to.equal(refund._id?.toString());

    refund = await findRefund(txHash);

    expect(refund).to.not.be.null;

    if (!refund) return;


    expect(refund.sequence).to.equal(account.sequence);
    expect(refund.status).to.be.oneOf([Status.SIGNED, Status.BROADCASTED]);
    debug("Refund signed");

    expect(refund.signatures.length).to.greaterThanOrEqual(2);

    debug("Refunding Order...");

    await sleep(3000);

    refund = await findRefund(txHash);

    expect(refund).to.not.be.null;

    if (!refund) return;

    expect(refund.status).to.equal(Status.BROADCASTED);
    expect(refund.transaction_hash).to.not.be.null;


    txHash = refund.transaction_hash.toLowerCase();

    tx = await findTransaction(txHash);

    expect(tx).to.not.be.null;

    if (!tx) return;
    debug("Refund transaction created");

    expect(tx.status).to.oneOf([Status.PENDING, Status.CONFIRMED]);
    expect(tx.from_address).to.equal(toHex);
    expect(tx.to_address).to.equal(fromHex);

    await sleep(3500);

    tx = await findTransaction(txHash);

    expect(tx).to.not.be.null;

    if (!tx) return;

    expect(tx.status).to.equal(Status.CONFIRMED);
    expect(tx.refund).to.not.be.null;

    if (!tx.refund) return;

    expect(tx.refund.toString()).to.equal(refund._id?.toString());

    refund = await findRefund(originTxHash);

    expect(refund).to.not.be.null;

    if (!refund) return;

    expect(refund.status).to.equal(Status.SUCCESS);
    expect(refund.transaction_hash.toLowerCase()).to.equal(tx.hash);

    expect(refund.transaction).to.not.be.null;
    expect(refund.transaction?.toString()).to.equal(tx._id?.toString());


    const afterBalance = await cosmos.getBalance(fromAddress);

    expect(afterBalance).to.equal(beforeBalance - BigInt(2) * POKT_TX_FEE);

    debug("Refund success");
  });

  it("should do multiple consecutive mints", async () => {
    debug("\nTesting -- should do multiple consecutive mints");

    const signer = await cosmos.signerPromise;
    const fromAddress = await cosmos.getAddress();
    const recipientAddress = await ethereum.getAddress(ethNetwork.chain_id);
    const toAddress = cosmosNetwork.multisig_address;

    const fromHex = cosmos.bech32ToHex(fromAddress);

    const amounts = [
      parseUnits("1", 6),
      parseUnits("2", 6),
      parseUnits("3", 6),
    ];

    const account = await cosmos.getAccount(fromAddress);

    expect(account).to.not.be.null;

    if (!account) return;


    const startNonce = account.sequence;

    const memo: MintMemo = {
      address: recipientAddress,
      chain_id: ethNetwork.chain_id.toString(),
    };

    const fromBeforeBalance = await cosmos.getBalance(fromAddress);
    const toBeforeBalance = await cosmos.getBalance(toAddress);

    debug("Sending transactions...");

    const sendTxs: Array<cosmos.CosmosTx> = [];

    for (let i = 0; i < amounts.length; i++) {
      const amount = amounts[i];

      const sendTx = await cosmos.sendPOKT(
        signer,
        toAddress,
        amount.toString(),
        JSON.stringify(memo),
        POKT_TX_FEE.toString(),
      );

      expect(sendTx).to.not.be.null;

      if (!sendTx) return;

      debug(`Transaction ${i} sent: `, sendTx.hash);

      sendTxs.push(sendTx);
    }

    expect(sendTxs).to.not.be.null;
    expect(sendTxs.length).to.equal(amounts.length);

    if (!sendTxs) return;

    const fromAfterBalance = await cosmos.getBalance(fromAddress);
    const toAfterBalance = await cosmos.getBalance(toAddress);

    const totalAmount = amounts.reduce(
      (total, amount) => total + amount,
      BigInt(0),
    );

    expect(fromAfterBalance).to.equal(
      fromBeforeBalance - totalAmount - BigInt(amounts.length) * POKT_TX_FEE,
    );
    expect(toAfterBalance).to.equal(toBeforeBalance + totalAmount);

    debug(`Waiting for messages to be signed...`);

    await sleep(6000);

    const noncesToSee = sendTxs.map((_, i) => (i + startNonce));

    const sortedMessages: Array<Message | null> = [null, null, null];

    for (let i = 0; i < sendTxs.length; i++) {
      const sendTx = sendTxs[i];

      if (!sendTx) return;

      const message = await findMessage(sendTx.hash);

      const txHash = "0x" + sendTx.hash.toLowerCase();

      expect(message).to.not.be.null;

      if (!message) return;

      expect(message.origin_transaction_hash).to.equal(txHash);

      const account = await cosmos.getAccount(fromAddress);

      expect(account).to.not.be.null;

      if (!account) return;

      expect(message.content.version).to.equal(cosmos.HYPERLANE_VERSION);
      expect(message.content.nonce).to.be.oneOf(noncesToSee);
      expect(message.content.origin_domain).to.equal(cosmos.CHAIN_DOMAIN);
      expect(message.content.sender).to.equal(fromHex);
      expect(message.content.destination_domain).to.equal(ethNetwork.chain_id);
      expect(message.content.recipient).to.equal(ethNetwork.mint_controller_address.toLowerCase());
      expect(message.content.message_body.sender_address).to.equal(fromHex);
      expect(message.content.message_body.recipient_address).to.equal(recipientAddress.toLowerCase());
      expect(message.content.message_body.amount.toString()).to.be.oneOf(amounts.map((a) => a.toString()));
      expect(message.signatures.length).to.be.greaterThanOrEqual(2);
      expect(message.status).to.equal(Status.SIGNED);
      debug(`Message ${i} signed`);

      const nonce = message.content.nonce;
      noncesToSee.splice(noncesToSee.indexOf(nonce), 1);

      const sortedIndex = nonce - startNonce;
      sortedMessages[sortedIndex] = message;
    }

    const fulfillmentTxs: Array<TransactionReceipt> = [];

    for (let i = 0; i < sortedMessages.length; i++) {
      const message = sortedMessages[i];
      expect(message).to.not.be.null;

      if (!message) return;

      debug(`Fulfilling Order ${i}...`);


      const beforeWPOKTBalance = await ethereum.getWPOKTBalance(
        ethNetwork.chain_id,
        recipientAddress
      );

      const messageBytes = encodeMessage(message.content);

      const metadata = concatHex(message.signatures.map((s) => s.signature));

      const fulfillmentTx = await ethereum.fulfillOrder(ethNetwork.chain_id, metadata, messageBytes);

      expect(fulfillmentTx).to.not.be.null;

      if (!fulfillmentTx) return;
      debug(`Fulfilled ${i}: `, fulfillmentTx.transactionHash);

      fulfillmentTxs.push(fulfillmentTx);

      const afterWPOKTBalance =
        await ethereum.getWPOKTBalance(ethNetwork.chain_id, recipientAddress);

      expect(afterWPOKTBalance).to.equal(
        beforeWPOKTBalance + BigInt(message.content.message_body.amount),
      );
    }

    await sleep(5000);

    await Promise.all(
      sortedMessages.map(async (oldMessage, i) => {
        expect(oldMessage).to.not.be.null;
        if (!oldMessage) return;

        const message = await findMessage(oldMessage.origin_transaction_hash);

        expect(message).to.not.be.null;

        if (!message) return;

        expect(message.status).to.equal(Status.SUCCESS);

        const fulfillmentTx = fulfillmentTxs[i];

        expect(message.transaction_hash.toLowerCase()).to.equal(
          fulfillmentTx.transactionHash.toLowerCase(),
        );
        debug(`Message ${i} success`);
      }),
    );
  });

};
