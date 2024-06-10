import { DirectSecp256k1HdWallet, OfflineDirectSigner } from "@cosmjs/proto-signing";
import { IndexedTx, SigningStargateClient, StargateClient } from "@cosmjs/stargate";
import { config } from "./config";
import { parseUnits } from "viem";
import { sleep } from "./helpers";


const FAUCET_MNEMONIC = "baby advance work soap slow exclude blur humble lucky rough teach wide chuckle captain rack laundry butter main very cannon donate armor dress follow";

const PREFIX = config.cosmos_network.bech32_prefix;

const RPC_ENDPOINT = config.cosmos_network.rpc_url;

const DENOM = config.cosmos_network.coin_denom;


const getAddress = async (): Promise<string> => {
  const signer = await signerPromise;
  const [account] = await signer.getAccounts();
  return account.address;
};


const getBalance = async (address: string): Promise<bigint> => {
  const client = await StargateClient.connect(RPC_ENDPOINT);

  const balances = await client.getAllBalances(address);

  const balance = balances.find((balance) => balance.denom === DENOM);

  return balance ? BigInt(balance.amount) : BigInt(0);
};

const POLL_INTERVAL = 1000;

const pollForTransaction = async (
  txHash: string
): Promise<IndexedTx | null> => {
  const client = await StargateClient.connect(RPC_ENDPOINT);

  let polls = 0;
  while (polls < 5) {
    try {
      const tx = await client.getTx(txHash);
      return tx;
    } catch {
      // do nothing
    } finally {
      await sleep(POLL_INTERVAL);
    }
  }
  return null;
};


const sendPOKT = async (
  signer: OfflineDirectSigner,
  recipient: string,
  amount: string,
  memo: string = "",
  feeAmount: string = ""
): Promise<IndexedTx | null> => {

  const client = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, signer);

  const amountFinal = {
    denom: DENOM, // Replace with your blockchain's denomination
    amount: amount,
  };
  const fee = {
    amount: feeAmount ? [{ denom: DENOM, amount: feeAmount }] : [], // Fee in uatom
    gas: "200000", // Gas limit
  };

  const [firstAccount] = await signer.getAccounts();

  const result = await client.sendTokens(firstAccount.address, recipient, [amountFinal], fee, memo);

  const tx = await pollForTransaction(result.transactionHash);

  return tx;
};

const signerPromise = (async (): Promise<DirectSecp256k1HdWallet> => {

  const faucetWallet = await DirectSecp256k1HdWallet.fromMnemonic(FAUCET_MNEMONIC, { prefix: PREFIX });

  const newSigner = await DirectSecp256k1HdWallet.generate(12, { prefix: PREFIX });

  const [newAccount] = await newSigner.getAccounts();

  await sendPOKT(
    faucetWallet,
    newAccount.address,
    parseUnits("100", 6).toString(),
    "init"
  );

  return newSigner;
})();

export default {
  getBalance,
  getAddress,
  sendPOKT,
  signerPromise,
  getTransction: pollForTransaction,
};
