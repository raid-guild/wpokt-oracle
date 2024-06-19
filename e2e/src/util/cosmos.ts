import {
  DirectSecp256k1HdWallet,
  OfflineDirectSigner,
} from "@cosmjs/proto-signing";
import {
  Account,
  Event,
  IndexedTx,
  SigningStargateClient,
  StargateClient,
} from "@cosmjs/stargate";
import { config } from "./config";
import {
  parseUnits,
  Hex,
} from "viem";
import { sleep } from "./helpers";
import { fromBech32 } from "@cosmjs/encoding";
import { keccak256 } from "viem";

function getChainDomain(chainId: string) {
  const chainHash = keccak256(new TextEncoder().encode(chainId));
  const chainDomain = BigInt(chainHash);
  return Number(chainDomain & BigInt(0xffffffff)); // Convert to uint32
}

const FAUCET_MNEMONIC =
  "baby advance work soap slow exclude blur humble lucky rough teach wide chuckle captain rack laundry butter main very cannon donate armor dress follow";

const PREFIX = config.cosmos_network.bech32_prefix;

const RPC_ENDPOINT = config.cosmos_network.rpc_url;

const DENOM = config.cosmos_network.coin_denom;

export const getAddress = async (): Promise<string> => {
  const signer = await getSigner();
  const [account] = await signer.getAccounts();
  return account.address;
};

export const getAccount = async (address: string): Promise<Account | null> => {
  const client = await StargateClient.connect(RPC_ENDPOINT);

  const account = await client.getAccount(address);

  return account;
};

export function bech32ToHex(bech32Address: string): Hex {
  const decoded = fromBech32(bech32Address);
  return ("0x" +
    Buffer.from(decoded.data).toString("hex").toLowerCase()) as Hex;
}

export const getBalance = async (address: string): Promise<bigint> => {
  const client = await StargateClient.connect(RPC_ENDPOINT);

  const balances = await client.getAllBalances(address);

  const balance = balances.find((balance) => balance.denom === DENOM);

  return balance ? BigInt(balance.amount) : BigInt(0);
};

const POLL_INTERVAL = 1000;

export const getTransaction = async (
  txHash: string,
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

export type CosmosTx = {
  readonly height: number;
  readonly txIndex: number;
  readonly hash: string;
  readonly code: number;
  readonly events: readonly Event[];
};

export const sendPOKT = async (
  signer: OfflineDirectSigner,
  recipient: string,
  amount: string,
  memo: string = "",
  feeAmount: string = "",
): Promise<CosmosTx | null> => {
  const client = await SigningStargateClient.connectWithSigner(
    RPC_ENDPOINT,
    signer,
  );

  const amountFinal = {
    denom: DENOM, // Replace with your blockchain's denomination
    amount: amount,
  };
  const fee = {
    amount:
      feeAmount && feeAmount != "0"
        ? [{ denom: DENOM, amount: feeAmount }]
        : [], // Fee in uatom
    gas: "200000", // Gas limit
  };

  const [firstAccount] = await signer.getAccounts();

  const result = await client.sendTokens(
    firstAccount.address,
    recipient,
    [amountFinal],
    fee,
    memo,
  );

  if (result.code !== 0) {
    return {
      ...result,
      hash: result.transactionHash,
    };
  }

  const tx = await getTransaction(result.transactionHash);

  return tx;
};

let signer: DirectSecp256k1HdWallet | null = null;

export const getSigner = async (): Promise<DirectSecp256k1HdWallet> => {

  if (!signer) {

    const faucetWallet = await DirectSecp256k1HdWallet.fromMnemonic(
      FAUCET_MNEMONIC,
      { prefix: PREFIX },
    );

    signer = await DirectSecp256k1HdWallet.generate(12, { prefix: PREFIX });

    const [newAccount] = await signer.getAccounts();

    await sendPOKT(
      faucetWallet,
      newAccount.address,
      parseUnits("9999", 6).toString(),
      "init"
    );

  }

  return signer;
}

export const CHAIN_DOMAIN = getChainDomain(config.cosmos_network.chain_id);

export const TX_FEE = BigInt(config.cosmos_network.tx_fee);
