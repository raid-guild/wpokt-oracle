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
} from 'viem';

import { AccountFactoryAbi, Multicall3Abi } from './abis';
import { AccountFactory } from './config';
import { getPublicClient } from './ethereum';

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

  const publicClient = getPublicClient(walletClient.chain.id);

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
  args: Array<WriteParams> | WriteParams,
): Promise<Hex> => {
  if (!Array.isArray(args)) {
    args = [args] as Array<WriteParams>;
  }

  const account = await getAccount(walletClient);

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

  const to = walletClient.chain?.contracts?.multicall3
    ?.address as Hex;

  if (!to) {
    throw new Error('Multicall contract address is not found');
  }

  const value = BigInt(0); // no payable amount

  const operation = BigInt(1); // delegatecall

  return walletClient.writeContract({
    chain: walletClient.chain as Chain,
    account: walletClient.account as Account,
    address: account,
    abi: parseAbi([
      'function execute(address to, uint256 value, bytes calldata data, uint256 operation) external',
    ]),
    functionName: 'execute',
    args: [to, value, data, operation],
  });
};
