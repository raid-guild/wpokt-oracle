import { concatHex, parseUnits } from "viem";
import * as ethereum from "../util/ethereum";
import * as cosmos from "../util/cosmos";
import { expect } from "chai";
import { sleep, debug } from "../util/helpers";
import { config } from "../util/config";
import { MintMemo, Status } from "../types";
import {
  findMessageByMessageID,
  findMessagesByTxHash,
  findTransaction,
} from "../util/mongodb";
import { fulfillSignedMessage } from "./helpers/fulfill";
import * as multi from "../util/account";
import { encodeMessage } from "../util/message";

const POKT_TX_FEE = cosmos.TX_FEE;

export const multipleBridgesFlow = async () => {
  const ethNetworkOne = config.ethereum_networks[0];
  const ethNetworkTwo = config.ethereum_networks[1];
  const cosmosNetwork = config.cosmos_network;

  it("should do multiple initiates in a single tx", async () => {
    debug(
      "\nTesting -- should do multiple initiates in a single tx"
    );

    const signer = await cosmos.getSigner();

    const ethSignerOne = await ethereum.getWallet(ethNetworkOne.chain_id);
    const ethAddressOne = ethSignerOne.account.address;
    debug("Eth Address One: ", ethAddressOne);
    const ethAccountOne = await multi.getAccount(ethSignerOne);
    debug("Eth Account One: ", ethAccountOne);

    const ethSignerTwo = await ethereum.getWallet(ethNetworkTwo.chain_id);
    const ethAddressTwo = ethSignerTwo.account.address;
    debug("Eth Address Two: ", ethAddressTwo);
    const ethAccountTwo = await multi.getAccount(ethSignerTwo);
    debug("Eth Account Two: ", ethAccountTwo);

    {
      debug("Bridging from Cosmos to Ethereum");

      const toAddress = cosmosNetwork.multisig_address;
      const amount = parseUnits("100", 6);

      const memo: MintMemo = {
        address: ethAccountOne,
        chain_id: ethNetworkOne.chain_id.toString(),
      };

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

      debug("Waiting for transaction to be confirmed...");
      await sleep(5000);

      const tx = await findTransaction(sendTx.hash, cosmosNetwork.chain_id);

      expect(tx).to.not.be.null;

      if (!tx) return;

      expect(tx.status).to.equal(Status.CONFIRMED);
      debug("Transaction confirmed");

      let messages = await findMessagesByTxHash(sendTx.hash);

      expect(messages).to.not.be.null;
      expect(messages.length).to.equal(1);

      let message = messages[0];

      expect(message).to.not.be.null;

      if (!message) return;
      debug("Message created");

      debug("Waiting for message to be signed...");
      await sleep(3500);

      await fulfillSignedMessage(message.message_id);

      const ethAccountOneBalance = await ethereum.getWPOKTBalance(ethNetworkOne.chain_id, ethAccountOne);

      expect(ethAccountOneBalance).to.equal(amount);

      debug("Eth Account One Balance has sufficient funds");
    }

    const beforeCosmosBalance = await cosmos.getBalance(await cosmos.getAddress());
    const beforeEthAddressTwoBalance = await ethereum.getWPOKTBalance(ethNetworkTwo.chain_id, ethAddressTwo);
    const beforeEthAccountTwoBalance = await ethereum.getWPOKTBalance(ethNetworkTwo.chain_id, ethAccountTwo);


    const cosmosHex = cosmos.bech32ToHex(await cosmos.getAddress())

    const orders: multi.InitiateParams[] = [
      {
        destinationDomain: cosmos.CHAIN_DOMAIN,
        amount: parseUnits("25", 6),
        recipientAddress: cosmosHex,
      },
      {
        destinationDomain: ethNetworkTwo.chain_id,
        amount: parseUnits("30", 6),
        recipientAddress: ethAddressTwo,
      },
      {
        destinationDomain: ethNetworkTwo.chain_id,
        amount: parseUnits("45", 6),
        recipientAddress: ethAccountTwo,
      },
    ];

    debug("Initiating multiple bridges...");

    const receipt = await multi.initiateMultiOrder(ethSignerOne, orders);

    expect(receipt).to.not.be.null;

    if (!receipt) return;

    const txHash = receipt.transactionHash;
    debug("Transaction sent: ", txHash);

    debug("Waiting for transaction to be confirmed...");
    await sleep(7000);

    const tx = await findTransaction(txHash, ethNetworkOne.chain_id);

    expect(tx).to.not.be.null;

    if (!tx) return;

    expect(tx.status).to.equal(Status.CONFIRMED);
    debug("Transaction confirmed");

    let messages = await findMessagesByTxHash(txHash);

    expect(messages).to.not.be.null;
    expect(messages.length).to.equal(3);

    debug("Messages created");

    debug("Waiting for message to be signed...");
    await sleep(7000);

    for (let i = 0; i < messages.length; i++) {
      const message = await findMessageByMessageID(messages[i].message_id);

      expect(message).to.not.be.null;

      if (!message) return;


      const destinationDomain = message.content.destination_domain;

      if (destinationDomain.toNumber() !== cosmos.CHAIN_DOMAIN) {
        expect(message.status).to.equal(Status.SIGNED);
        debug(`Fulfilling Ethereum message ${i}: `, message.message_id);
        const fulfilledMessage = await fulfillSignedMessage(message.message_id, false);
        expect(fulfilledMessage).to.not.be.null;

        if (!fulfilledMessage) return;
        expect(fulfilledMessage.status).to.equal(Status.SUCCESS);
        debug(`Eth Message ${i} fulfilled`);
      } else {
        expect(message.status).to.oneOf([Status.SIGNED, Status.BROADCASTED, Status.SUCCESS]);
        debug("Waiting for cosmos message to be successful...");
        await sleep(5000);

        const fulfilledMessage = await findMessageByMessageID(message.message_id);
        expect(fulfilledMessage).to.not.be.null;

        if (!fulfilledMessage) return;
        expect(fulfilledMessage.status).to.equal(Status.SUCCESS);
        debug(`Cosmos Message ${i} fulfilled`);
      }
    }


    const afterCosmosBalance = await cosmos.getBalance(await cosmos.getAddress());
    const afterEthAddressTwoBalance = await ethereum.getWPOKTBalance(ethNetworkTwo.chain_id, ethAddressTwo);
    const afterEthAccountTwoBalance = await ethereum.getWPOKTBalance(ethNetworkTwo.chain_id, ethAccountTwo);

    expect(afterCosmosBalance).to.equal(beforeCosmosBalance + orders[0].amount - POKT_TX_FEE);
    expect(afterEthAddressTwoBalance).to.equal(beforeEthAddressTwoBalance + orders[1].amount);
    expect(afterEthAccountTwoBalance).to.equal(beforeEthAccountTwoBalance + orders[2].amount);

    debug("Multiple bridges successful");
  });

  it("should do multiple fulfills in a single tx", async () => {
    debug(
      "\nTesting -- should do multiple fulfills in a single tx"
    );

    const signer = await cosmos.getSigner();

    const ethSignerOne = await ethereum.getWallet(ethNetworkOne.chain_id);
    const ethAddressOne = ethSignerOne.account.address;
    debug("Eth Address One: ", ethAddressOne);
    const ethAccountOne = await multi.getAccount(ethSignerOne);
    debug("Eth Account One: ", ethAccountOne);

    const ethSignerTwo = await ethereum.getWallet(ethNetworkTwo.chain_id);
    const ethAddressTwo = ethSignerTwo.account.address;
    debug("Eth Address Two: ", ethAddressTwo);
    const ethAccountTwo = await multi.getAccount(ethSignerTwo);
    debug("Eth Account Two: ", ethAccountTwo);

    {
      debug("Bridging from Cosmos to Ethereum");

      const toAddress = cosmosNetwork.multisig_address;
      const amount = parseUnits("100", 6);

      const memo: MintMemo = {
        address: ethAccountOne,
        chain_id: ethNetworkOne.chain_id.toString(),
      };

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

      debug("Waiting for transaction to be confirmed...");
      await sleep(5000);

      const tx = await findTransaction(sendTx.hash, cosmosNetwork.chain_id);

      expect(tx).to.not.be.null;

      if (!tx) return;

      expect(tx.status).to.equal(Status.CONFIRMED);
      debug("Transaction confirmed");

      let messages = await findMessagesByTxHash(sendTx.hash);

      expect(messages).to.not.be.null;
      expect(messages.length).to.equal(1);

      let message = messages[0];

      expect(message).to.not.be.null;

      if (!message) return;
      debug("Message created");

      debug("Waiting for message to be signed...");
      await sleep(3500);

      await fulfillSignedMessage(message.message_id, false);

      const ethAccountOneBalance = await ethereum.getWPOKTBalance(ethNetworkOne.chain_id, ethAccountOne);

      expect(ethAccountOneBalance >= amount).to.be.true;

      debug("Eth Account One Balance has sufficient funds");
    }

    const beforeEthAddressTwoBalance = await ethereum.getWPOKTBalance(ethNetworkTwo.chain_id, ethAddressTwo);
    const beforeEthAccountTwoBalance = await ethereum.getWPOKTBalance(ethNetworkTwo.chain_id, ethAccountTwo);

    const orders: multi.InitiateParams[] = [
      {
        destinationDomain: ethNetworkTwo.chain_id,
        amount: parseUnits("25", 6),
        recipientAddress: ethAddressTwo,
      },
      {
        destinationDomain: ethNetworkTwo.chain_id,
        amount: parseUnits("30", 6),
        recipientAddress: ethAddressTwo,
      },
      {
        destinationDomain: ethNetworkTwo.chain_id,
        amount: parseUnits("45", 6),
        recipientAddress: ethAccountTwo,
      },
    ];

    debug("Initiating multiple bridges...");

    const receipt = await multi.initiateMultiOrder(ethSignerOne, orders);

    expect(receipt).to.not.be.null;

    if (!receipt) return;

    const txHash = receipt.transactionHash;
    debug("Transaction sent: ", txHash);

    debug("Waiting for transaction to be confirmed...");
    await sleep(7000);

    const tx = await findTransaction(txHash, ethNetworkOne.chain_id);

    expect(tx).to.not.be.null;

    if (!tx) return;

    expect(tx.status).to.equal(Status.CONFIRMED);
    debug("Transaction confirmed");

    const messages = await findMessagesByTxHash(txHash);

    expect(messages).to.not.be.null;
    expect(messages.length).to.equal(3);

    debug("Messages created");

    debug("Waiting for message to be signed...");
    await sleep(9000);

    const fulfills: multi.FulfillParams[] = [];

    for (let i = 0; i < messages.length; i++) {
      const message = await findMessageByMessageID(messages[i].message_id);

      expect(message).to.not.be.null;

      if (!message) return;

      expect(message.status).to.equal(Status.SIGNED);
      expect(message.signatures.length).to.equal(3);

      const messageBytes = encodeMessage(message.content);

      const metadata = concatHex(message.signatures.map((s) => s.signature));

      fulfills.push({
        metadata: metadata,
        message: messageBytes,
      });
    }

    debug("Fulfilling multiple bridges...");


    const fulfillReceipt = await multi.fulfillMultiOrder(ethSignerTwo, fulfills);

    expect(fulfillReceipt).to.not.be.null;

    if (!fulfillReceipt) return;

    const fulfillTxHash = fulfillReceipt.transactionHash;

    debug("Transaction sent: ", fulfillTxHash);

    debug("Waiting for transaction to be confirmed...");

    await sleep(7000);

    const fulfillTx = await findTransaction(fulfillTxHash, ethNetworkTwo.chain_id);

    expect(fulfillTx).to.not.be.null;

    if (!fulfillTx) return;

    expect(fulfillTx.status).to.equal(Status.CONFIRMED);

    debug("Transaction confirmed");


    expect(fulfillTx.messages.length).to.equal(3);

    for (let i = 0; i < messages.length; i++) {
      expect(messages[i]._id.toString()).to.be.oneOf(fulfillTx.messages.map((m) => m.toString()));
    }

    debug("Transaction confirmed");


    const afterEthAddressTwoBalance = await ethereum.getWPOKTBalance(ethNetworkTwo.chain_id, ethAddressTwo);
    const afterEthAccountTwoBalance = await ethereum.getWPOKTBalance(ethNetworkTwo.chain_id, ethAccountTwo);

    expect(afterEthAddressTwoBalance).to.equal(beforeEthAddressTwoBalance + orders[0].amount + orders[1].amount);
    expect(afterEthAccountTwoBalance).to.equal(beforeEthAccountTwoBalance + orders[2].amount);

    debug("Multiple bridges successful");
  });

};
