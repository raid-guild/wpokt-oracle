import {
  Account,
  Hex,
  TransactionReceipt,
  Transport,
  WalletClient,
  createPublicClient,
  createWalletClient,
  decodeEventLog,
  defineChain,
  encodeEventTopics,
  http,
  // parseUnits,
} from "viem";
import {
  // generatePrivateKey, 
  privateKeyToAccount
} from "viem/accounts";
import { Chain } from "viem/chains";
import { EthereumNetworkConfig, config } from "./config";
import { MintControllerAbi, OmniTokenAbi } from "./abis";


const DEFAULT_PRIVATE_KEY = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80";

const createChain = (ethNetwork: EthereumNetworkConfig) => defineChain({
  id: ethNetwork.chain_id,
  name: ethNetwork.chain_name,
  nativeCurrency: {
    decimals: 18,
    name: 'Ether',
    symbol: 'ETH',
  },
  rpcUrls: {
    default: { http: [ethNetwork.rpc_url] },
  },
})

const chains: Record<number, Chain> = config.ethereum_networks.reduce(
  (acc, network) => ({
    ...acc,
    [network.chain_id]: createChain(network),
  }),
  {}
);

const networkConfig: Record<number, EthereumNetworkConfig> = config.ethereum_networks.reduce(
  (acc, network) => ({
    ...acc,
    [network.chain_id]: network,
  }),
  {}
);

const defaultWalletClient: (chain_id: number) => WalletClient<Transport, Chain, Account> = (chain_id: number) =>
  createWalletClient({
    account: privateKeyToAccount(DEFAULT_PRIVATE_KEY),
    chain: chains[chain_id],
    transport: http(),
  });

const publicClient: (chain_id: number) => ReturnType<typeof createPublicClient> = (chain_id: number) => createPublicClient({
  chain: chains[chain_id],
  transport: http(),
});

export const getBalance = async (chain_id: number, address: Hex): Promise<bigint> => {
  const balance = await publicClient(chain_id).getBalance({
    address,
  });
  return balance;
};

export const getWPOKTBalance = async (chain_id: number, address: Hex): Promise<bigint> => {
  const tokenAddress = networkConfig[chain_id].omni_token_address as Hex;
  const balance = await publicClient(chain_id).readContract({
    address: tokenAddress,
    abi: OmniTokenAbi,
    functionName: "balanceOf",
    args: [address],
  });

  return balance as bigint;
};

export const sendETH = async (
  wallet: WalletClient<Transport, Chain, Account>,
  recipient: Hex,
  amount: bigint
): Promise<TransactionReceipt> => {
  const hash = await wallet.sendTransaction({
    to: recipient,
    value: amount,
  });
  const receipt = await publicClient(wallet.chain.id).waitForTransactionReceipt({ hash });
  return receipt;
};

export const sendWPOKT = async (
  wallet: WalletClient<Transport, Chain, Account>,
  recipient: Hex,
  amount: bigint
): Promise<TransactionReceipt> => {
  const chain_id = wallet.chain.id;
  const tokenAddress = networkConfig[chain_id].omni_token_address as Hex;
  const hash = await wallet.writeContract({
    address: tokenAddress,
    abi: OmniTokenAbi,
    functionName: "transfer",
    args: [recipient, amount],
  });
  const receipt = await publicClient(chain_id).waitForTransactionReceipt({ hash });
  return receipt;
};

export const getWallet: (chain_id: number) => Promise<WalletClient<Transport, Chain, Account>> =
  async (chain_id: number) => {
    // const pKey = generatePrivateKey();
    // const walletClient = createWalletClient({
    //   account: privateKeyToAccount(pKey),
    //   chain: chains[chain_id],
    //   transport: http(),
    // });
    //
    // await sendETH(
    //   defaultWalletClient(chain_id),
    //   walletClient.account.address,
    //   parseUnits("10", 18)
    // );
    const walletClient = defaultWalletClient(chain_id);

    return walletClient;
  };

export const getAddress = async (chain_id: number): Promise<Hex> => {
  const wallet = await getWallet(chain_id);
  return wallet.account.address.toLowerCase() as Hex;
};


export const fulfillOrder = async (chain_id: number, metadata: Hex, message: Hex): Promise<TransactionReceipt> => {
  const wallet = await getWallet(chain_id);
  const hash = await wallet.writeContract({
    address: networkConfig[chain_id].mint_controller_address as Hex,
    abi: MintControllerAbi,
    functionName: "fulfillOrder",
    args: [metadata, message],
  });

  const receipt = await publicClient(chain_id).waitForTransactionReceipt({ hash });
  return receipt;
}

export type FulfillmentEvent = {
  orderId: Hex;
  message: Hex;
};

export const findFulfillmentEvent = (receipt: TransactionReceipt): FulfillmentEvent | null => {
  const eventTops = encodeEventTopics({
    abi: MintControllerAbi,
    eventName: "Fulfillment",
  });

  const event = receipt.logs.find((log) => log.topics[0] === eventTops[0]);

  if (!event) {
    return null;
  }

  const decodedLog = decodeEventLog({
    abi: MintControllerAbi,
    eventName: "Fulfillment",
    data: event.data,
    topics: event.topics,
  });

  return decodedLog.args as unknown as FulfillmentEvent;
};
