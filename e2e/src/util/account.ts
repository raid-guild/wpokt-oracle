import {
  Hex,
  Abi,
  Account,
  Chain,
  encodeFunctionData,
  parseAbi,
  WalletClient,
  WriteContractParameters,
  Transport,
  zeroAddress,
  PublicClient,
  TransactionReceipt,
  formatUnits,
} from 'viem';

import { AccountFactoryAbi, MintControllerAbi, Multicall3Abi, OmniTokenAbi } from './abis';
import { AccountFactory, MintController, Multicall3 } from './config';
import * as ethereum from './ethereum';
import { addressHexToBytes32, formatMessageBody } from './message';
import { debug } from './helpers';

type WriteParams = WriteContractParameters<
  Abi,
  string,
  readonly unknown[],
  Chain | undefined
>;

const accounts: Record<Hex, Hex> = {};

const createAndGetAccount = async (
  publicClient: PublicClient<Transport, Chain, Account>,
  walletClient: WalletClient<Transport, Chain, Account>,
): Promise<Hex> => {

  const { request, result } = await publicClient.simulateContract({
    chain: walletClient.chain as Chain,
    account: walletClient.account as Account,
    address: AccountFactory,
    abi: AccountFactoryAbi,
    functionName: 'getAccount',
    args: [],
  })

  const hash = await walletClient.writeContract(request)

  await publicClient.waitForTransactionReceipt({
    hash,
  });

  accounts[walletClient.account.address] = result as Hex

  return result as Hex
}

export const getAccount = async (
  walletClient: WalletClient<Transport, Chain, Account>,
): Promise<Hex> => {
  const address = walletClient.account.address;

  if (!!accounts[address]) {
    return accounts[address];
  }

  const publicClient = ethereum.getPublicClient(walletClient.chain.id);

  const account = await publicClient.readContract({
    address: AccountFactory,
    abi: AccountFactoryAbi,
    functionName: "accounts",
    args: [address],
  });

  if (account != zeroAddress) {
    accounts[address] = account as Hex;
    return account as Hex;
  }



  return createAndGetAccount(publicClient, walletClient);
}

export const executeAsAccount = async (
  walletClient: WalletClient<Transport, Chain, Account>,
  args: Array<WriteParams>,
): Promise<TransactionReceipt> => {

  const calls = args.map(arg => {
    const data = encodeFunctionData({
      abi: arg.abi,
      functionName: arg.functionName,
      args: arg.args,
    });

    return {
      target: arg.address,
      callData: data,
    };
  });

  const data = encodeFunctionData({
    abi: Multicall3Abi,
    functionName: 'aggregate',
    args: [calls],
  });

  const value = BigInt(0); // no payable amount

  const operation = BigInt(1); // delegatecall

  debug("Executing multicall...");

  const hash = await walletClient.writeContract({
    chain: walletClient.chain as Chain,
    account: walletClient.account as Account,
    address: await getAccount(walletClient),
    abi: parseAbi([
      'function execute(address to, uint256 value, bytes calldata data, uint256 operation) external',
    ]),
    functionName: 'execute',
    args: [Multicall3, value, data, operation],
  });

  debug("Multicall executed: ", hash);

  return ethereum.getPublicClient(walletClient.chain.id).waitForTransactionReceipt({ hash });
};

export type InitiateParams = {
  destinationDomain: number,
  recipientAddress: Hex,
  amount: bigint,
}

export const initiateMultiOrder = async (
  wallet: WalletClient<Transport, Chain, Account>,
  params: InitiateParams[],
): Promise<TransactionReceipt> => {

  const chain_id = wallet.chain.id;

  const totalAmount = BigInt(params.reduce((t, p) => t + p.amount, BigInt(0)));

  debug("Total amount: ", formatUnits(totalAmount, 6));

  const approveParams: WriteParams = {
    chain: wallet.chain,
    account: wallet.account as Account,
    address: ethereum.networkConfig[chain_id].omni_token_address as Hex,
    abi: OmniTokenAbi,
    functionName: "approve",
    args: [MintController, totalAmount],
  }

  const writeParams = [approveParams]

  const account = await getAccount(wallet);
  params.forEach((param) => {

    const { destinationDomain, recipientAddress, amount } = param;

    const messageBody = formatMessageBody(
      recipientAddress,
      amount,
      account,
    );

    const destMintControllerAddress = ethereum.getMintControllerAddress(destinationDomain);

    const args = [
      destinationDomain,
      addressHexToBytes32(destMintControllerAddress),
      messageBody,
    ];

    const initiateParams: WriteParams = {
      chain: wallet.chain,
      account: wallet.account as Account,
      address: MintController,
      abi: MintControllerAbi,
      functionName: "initiateOrder",
      args: args,
    };

    writeParams.push(initiateParams);
  })

  debug("Write params initialized");

  return executeAsAccount(wallet, writeParams);
};

export type FulfillParams = {
  metadata: Hex,
  message: Hex,
}

export const fulfillMultiOrder = async (
  wallet: WalletClient<Transport, Chain, Account>,
  params: FulfillParams[],
): Promise<TransactionReceipt> => {

  const writeParams: WriteParams[] = params.map((param) => ({
    chain: wallet.chain,
    account: wallet.account as Account,
    address: MintController,
    abi: MintControllerAbi,
    functionName: "fulfillOrder",
    args: [param.metadata, param.message],
  }));

  return executeAsAccount(wallet, writeParams);
}
